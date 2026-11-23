import http from 'node:http';
import type net from 'node:net';

/**
 * Fake upstream for the media lookup sources (AniList, later TMDB / Open Library / iTunes). One per
 * worker; the worker's backend is pointed at it with LOOKUP_*_URL (backend/internal/core/media).
 * The real backend code runs (GraphQL request, parsing, error mapping); only the third party is
 * replaced, so "AniList is down" is produced by making this server fail, not by faking a response.
 */
export interface AniListMedia {
  id: number;
  type?: 'ANIME' | 'MANGA';
  title: string;
  year?: number;
  studio?: string;
  episodes?: number;
  genres?: string[];
  description?: string;
}

export interface UpstreamRequest {
  source: 'anilist';
  body: { query: string; variables: Record<string, unknown> };
}

export class Upstream {
  private catalog: AniListMedia[] = [];
  private failures: { status: number; times: number }[] = [];
  readonly requests: UpstreamRequest[] = [];

  constructor(
    private server: http.Server,
    readonly url: string,
  ) {}

  /** Replace the AniList catalogue (matched by case-insensitive substring of the title). */
  setAniList(items: AniListMedia[]) {
    this.catalog = items;
  }

  /** Make the next `times` AniList requests answer with `status` (e.g. 500, 429). */
  failAniList(status = 500, times = 1) {
    this.failures.push({ status, times });
  }

  /** Back to an empty catalogue, no failures, no recorded requests. Run before every test. */
  reset() {
    this.catalog = [];
    this.failures = [];
    this.requests.length = 0;
  }

  async stop() {
    this.server.closeAllConnections();
    await new Promise<void>((r) => this.server.close(() => r()));
  }

  handle(req: http.IncomingMessage, res: http.ServerResponse, raw: string) {
    const json = (status: number, body: unknown) => {
      res.writeHead(status, { 'content-type': 'application/json' });
      res.end(JSON.stringify(body));
    };
    if (req.method !== 'POST' || req.url !== '/anilist') return json(404, { error: 'not faked' });
    const body = JSON.parse(raw);
    this.requests.push({ source: 'anilist', body });
    const fail = this.failures[0];
    if (fail) {
      if (--fail.times <= 0) this.failures.shift();
      return json(fail.status, { errors: [{ message: 'fake upstream failure', status: fail.status }] });
    }
    const vars = body.variables ?? {};
    if (typeof vars.id === 'number') {
      const m = this.catalog.find((c) => c.id === vars.id);
      return m ? json(200, { data: { Media: toGraph(m) } }) : json(200, { data: { Media: null } });
    }
    const q = String(vars.search ?? '').toLowerCase();
    const type = vars.type === 'MANGA' ? 'MANGA' : 'ANIME';
    const media = this.catalog
      .filter((c) => (c.type ?? 'ANIME') === type && c.title.toLowerCase().includes(q))
      .map(toGraph);
    json(200, { data: { Page: { media } } });
  }
}

function toGraph(m: AniListMedia) {
  return {
    id: m.id,
    type: m.type ?? 'ANIME',
    format: 'TV',
    title: { romaji: m.title, english: m.title },
    startDate: { year: m.year ?? null },
    coverImage: { extraLarge: '', large: '' },
    description: m.description ?? '',
    genres: m.genres ?? [],
    episodes: m.episodes ?? null,
    chapters: null,
    volumes: null,
    studios: { nodes: m.studio ? [{ name: m.studio }] : [] },
    staff: { edges: [], nodes: [] },
  };
}

export async function startUpstream(): Promise<Upstream> {
  let up!: Upstream;
  const server = http.createServer((req, res) => {
    let raw = '';
    req.on('data', (c) => (raw += c));
    req.on('end', () => up.handle(req, res, raw));
  });
  await new Promise<void>((r) => server.listen(0, '127.0.0.1', r));
  const { port } = server.address() as net.AddressInfo;
  up = new Upstream(server, `http://127.0.0.1:${port}`);
  return up;
}
