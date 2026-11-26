import fs from 'node:fs';
import path from 'node:path';
import type { FullConfig, FullResult, Reporter, TestCase, TestResult } from '@playwright/test/reporter';

/**
 * Writes test-results/failures.md: one short block per failed (or flaky) test, so an agent reads one
 * file instead of opening traces. Each block has the error, the failing line, the test's last API
 * calls and backend errors (from the backend.log the fixtures attach), and the artifact paths.
 * Nothing is written when everything passed.
 */
const strip = (s: string) => s.replace(/\u001b\[[0-9;]*m/g, '');

interface Entry {
  test: TestCase;
  result: TestResult;
  flaky: boolean;
}

export default class SummaryReporter implements Reporter {
  private outDir = 'test-results';
  private failed = new Map<string, Entry>();
  private passedAfterRetry = new Set<string>();

  onBegin(config: FullConfig): void {
    this.outDir = config.projects[0]?.outputDir ?? this.outDir;
  }

  onTestEnd(test: TestCase, result: TestResult): void {
    if (result.status === test.expectedStatus) {
      if (this.failed.has(test.id)) this.passedAfterRetry.add(test.id);
      return;
    }
    if (result.status === 'skipped' || result.status === 'interrupted') return;
    this.failed.set(test.id, { test, result, flaky: false });
  }

  onEnd(_result: FullResult): void {
    const file = path.join(this.outDir, 'failures.md');
    const entries = [...this.failed.values()].map((e) => ({ ...e, flaky: this.passedAfterRetry.has(e.test.id) }));
    if (!entries.length) {
      fs.rmSync(file, { force: true });
      return;
    }
    const failed = entries.filter((e) => !e.flaky).length;
    const flaky = entries.length - failed;
    const out = [`# e2e failures: ${failed} failed, ${flaky} flaky`, ''];
    for (const { test, result, flaky: isFlaky } of entries) out.push(...this.block(test, result, isFlaky));
    fs.mkdirSync(this.outDir, { recursive: true });
    fs.writeFileSync(file, out.join('\n'));
    console.log(`\ne2e: failure summary → ${path.relative(process.cwd(), file)}`);
  }

  private block(test: TestCase, result: TestResult, flaky: boolean): string[] {
    const project = test.parent.project()?.name ?? '';
    const title = test.titlePath().filter(Boolean).slice(1).join(' › ');
    const loc = `${path.relative(process.cwd(), test.location.file)}:${test.location.line}`;
    const lines = [`## ${flaky ? '[flaky] ' : ''}[${project}] ${title}`, '', `- spec: \`${loc}\``];
    const err = result.errors[0];
    if (err?.location && fs.existsSync(err.location.file)) lines.push(`- failed at: \`${path.relative(process.cwd(), err.location.file)}:${err.location.line}\``);
    lines.push(`- rerun: \`npx playwright test ${loc} --project=${project}\``);
    // The fixtures write backend.log / db.sql / db-changes.txt into the test's output dir;
    // Playwright also keeps hashed copies under attachments/. Point at the readable originals.
    const outDir = result.attachments.map((a) => a.path).find((p) => p && !p.includes(`${path.sep}attachments${path.sep}`));
    const file = (a: TestResult['attachments'][number]) => {
      const original = outDir && path.join(path.dirname(outDir), a.name);
      return original && fs.existsSync(original) ? original : a.path;
    };
    for (const a of result.attachments) {
      const p = file(a);
      if (p) lines.push(`- ${a.name}: \`${path.relative(process.cwd(), p)}\``);
    }
    if (err) {
      const msg = strip(err.message ?? '').split('\n').filter(Boolean).slice(0, 12).join('\n');
      lines.push('', '```', msg, '```');
      if (err.snippet) lines.push('', '```', strip(err.snippet), '```');
    }
    const changes = result.attachments.find((a) => a.name === 'db-changes.txt');
    const changesFile = changes && file(changes);
    if (changesFile && fs.existsSync(changesFile)) {
      lines.push('', 'Data this test changed (created, updated or deleted rows):', '```', fs.readFileSync(changesFile, 'utf8'), '```');
    }
    const log = result.attachments.find((a) => a.name === 'backend.log');
    const logFile = log && file(log);
    if (logFile && fs.existsSync(logFile)) {
      const text = fs.readFileSync(logFile, 'utf8').split('\n');
      const calls = text.filter((l) => l.includes('[GIN]') && !l.includes('/__e2e/')).slice(-6);
      const errors = text.filter((l) => /\bERROR\b|panic/.test(l)).slice(-5);
      if (calls.length) lines.push('', 'Last API calls:', '```', ...calls.map((l) => l.replace(/^\[GIN\] \S+ - /, '')), '```');
      if (errors.length) lines.push('', 'Backend errors:', '```', ...errors, '```');
    }
    lines.push('');
    return lines;
  }
}
