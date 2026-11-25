// `npm run ci:history`: the CI e2e record (written by ci.sh), and whether it meets the bar for
// making the step blocking (TASKS C2): the last 20 runs all green with no flaky tests.
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

const file = path.join(process.env.XDG_STATE_HOME ?? path.join(os.homedir(), '.local/state'), 'simple-track-e2e/ci-history.jsonl');
const WINDOW = 20;

type Run = { at: string; commit: string; exit: number; secs: number; passed: number; failed: number; flaky: number };
const runs: Run[] = fs.existsSync(file)
  ? fs.readFileSync(file, 'utf8').split('\n').filter(Boolean).map((l) => JSON.parse(l))
  : [];

for (const r of runs.slice(-30)) {
  const mark = r.exit === 0 && !r.flaky ? 'ok  ' : r.exit === 0 ? 'FLKY' : 'FAIL';
  console.log(`${mark} ${r.at.slice(0, 16)} ${r.commit} ${r.passed ?? '?'} passed, ${r.failed ?? '?'} failed, ${r.flaky ?? '?'} flaky, ${r.secs}s`);
}
const recent = runs.slice(-WINDOW);
const clean = recent.filter((r) => r.exit === 0 && !r.flaky).length;
console.log(`\n${runs.length} runs recorded (${file}).`);
console.log(
  recent.length >= WINDOW && clean === WINDOW
    ? `Last ${WINDOW} runs clean: ready to set E2E_BLOCKING=1 in .cicd.yaml.`
    : `Not ready to block: ${clean}/${WINDOW} of the last ${Math.min(WINDOW, recent.length)} runs clean (needs ${WINDOW}/${WINDOW}).`,
);
