import type { APIRequestContext } from '@playwright/test';

/**
 * Client for the backend's test-only /__e2e endpoints (backend/internal/e2e). One per test, bound
 * to that worker's backend. See e2e/README.md, "Trigger the real cause".
 */

export type FaultOp = 'create' | 'query' | 'update' | 'delete' | 'row' | 'raw';

export interface FaultSpec {
  /** DB table, e.g. "logged_sets". Validated when armed. */
  table: string;
  op: FaultOp;
  /** Fail only the next n matching statements (default: every one). */
  times?: number;
  /** Return a Postgres error with this SQLSTATE, e.g. "23505" (unique_violation), "40001". */
  code?: string;
  message?: string;
  /** Don't fail the test if this fault never fires. */
  optional?: boolean;
}

export interface HoldSpec {
  /** "before" (default): wait before the handler runs. "after": the handler commits, the response waits. */
  phase?: 'before' | 'after';
  /** Safety net: auto-release after this long (default 15s). A timed-out hold fails the test. */
  timeoutMs?: number;
}

export class Gate {
  constructor(
    private readonly client: E2E,
    readonly id: string,
    readonly label: string,
  ) {}

  /** Resolves once the held request has reached the server (phase after: once it has committed). */
  async arrived(timeoutMs = 10_000): Promise<void> {
    const deadline = Date.now() + timeoutMs;
    while (Date.now() < deadline) {
      const r = await this.client.get(`/__e2e/holds/${this.id}?wait=${Math.min(1000, deadline - Date.now())}`);
      if ((await r.json()).arrived) return;
    }
    throw new Error(`hold ${this.label}: no matching request arrived within ${timeoutMs}ms`);
  }

  /** Let the request continue normally. */
  release(): Promise<void> {
    return this.client.decide(this.id, 'release');
  }

  /** Arm a DB fault (times: 1 by default), then let the request run into it. Phase before only. */
  fail(fault: FaultSpec): Promise<void> {
    this.client.expectServerErrors = true;
    return this.client.decide(this.id, 'fail', { times: 1, ...fault });
  }

  /** Close the connection without a response: the browser sees a network error. */
  drop(): Promise<void> {
    this.client.expectNetworkErrors = true;
    return this.client.decide(this.id, 'drop');
  }
}

export class E2E {
  /** Set when the test arms something that makes 5xx responses expected (see fixtures' console guard). */
  expectServerErrors = false;
  /** Set when the test drops connections. */
  expectNetworkErrors = false;
  readonly optionalFaults = new Set<string>();

  constructor(
    private readonly request: APIRequestContext,
    readonly backendUrl: string,
  ) {}

  async get(path: string) {
    const r = await this.request.get(this.backendUrl + path);
    if (!r.ok()) throw new Error(`GET ${path}: ${r.status()} ${await r.text()}`);
    return r;
  }

  private async post(path: string, data?: unknown) {
    const r = await this.request.post(this.backendUrl + path, { data });
    if (!r.ok()) throw new Error(`POST ${path}: ${r.status()} ${await r.text()}`);
    return r;
  }

  /** Truncate + reseed, reset the clock, clear faults and holds. The fixtures do this before every test. */
  async reset(): Promise<void> {
    await this.post('/__e2e/reset');
  }

  /** Add a data profile on top of the seed (backend/internal/e2e/seed: "large"). Prefer the `seed` option. */
  async seedProfile(profile: 'large'): Promise<void> {
    await this.post(`/__e2e/seed/${profile}`);
  }

  /** Move the server's clock (it keeps ticking from there). */
  async setClock(time: Date): Promise<void> {
    await this.post('/__e2e/clock', { time: time.toISOString() });
  }

  /** Make matching DB statements fail; the real handler turns that into its real response. */
  async fault(spec: FaultSpec): Promise<string> {
    const { optional, ...body } = spec;
    const r = await this.post('/__e2e/faults', body);
    const { id } = await r.json();
    if (optional) this.optionalFaults.add(id);
    this.expectServerErrors = true;
    return id;
  }

  /**
   * Hold the next request matching `route` ("POST /workout/sets", backend path without /api; a
   * template like "/diet/foods/:id" or a concrete path).
   */
  async hold(route: string, spec: HoldSpec = {}): Promise<Gate> {
    const [method, path] = route.trim().split(/\s+/);
    const r = await this.post('/__e2e/holds', { method, path: path.replace(/^\/api(?=\/)/, ''), ...spec });
    return new Gate(this, (await r.json()).id, route);
  }

  async decide(id: string, action: string, body?: unknown): Promise<void> {
    await this.post(`/__e2e/holds/${id}/${action}`, body);
  }

  async state(): Promise<{
    faults: { id: string; table: string; op: string; hits: number }[];
    holds: { id: string; method: string; path: string; arrived: boolean; decided: boolean; timedOut: boolean }[];
  }> {
    return (await this.get('/__e2e/state')).json();
  }

  /** Problems that should fail an otherwise passing test: unused faults, unused or stuck holds. */
  async leftovers(): Promise<string[]> {
    const s = await this.state();
    const out: string[] = [];
    for (const f of s.faults) {
      if (f.hits === 0 && !this.optionalFaults.has(f.id)) out.push(`fault ${f.op} ${f.table} was armed but never hit`);
    }
    for (const h of s.holds) {
      if (h.timedOut) out.push(`hold ${h.method} ${h.path} timed out (never released, failed or dropped)`);
      else if (!h.arrived) out.push(`hold ${h.method} ${h.path} was armed but no request matched it`);
      else if (!h.decided) out.push(`hold ${h.method} ${h.path} was never released, failed or dropped`);
    }
    return out;
  }
}
