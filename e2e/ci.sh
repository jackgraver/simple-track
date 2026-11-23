#!/usr/bin/env bash
# The j-cicd `e2e` step (.cicd.yaml). Runs the whole suite with CI settings (retries 1,
# --fail-on-flaky-tests), records one line per run in the history file below and keeps the report
# of a failed run, since the pipeline's snapshot checkout is thrown away.
#
# Non-blocking until E2E_BLOCKING=1 (TASKS C2): a red suite is reported but the step exits 0.
# `npm run ci:history` shows the record that decides when to flip it.
set -uo pipefail
cd "$(dirname "$0")"

STATE="${XDG_STATE_HOME:-$HOME/.local/state}/simple-track-e2e"
HISTORY="$STATE/ci-history.jsonl"
mkdir -p "$STATE"

npm ci --no-audit --no-fund --loglevel=error || exit 1
npx playwright install chromium >/dev/null || exit 1

start=$(date +%s)
CI=1 E2E_WORKERS="${E2E_WORKERS:-6}" npx playwright test
status=$?
secs=$(( $(date +%s) - start ))

commit="$(git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)"
node -e '
  const fs = require("fs");
  const [file, commit, status, secs] = process.argv.slice(1);
  let s = {};
  try { s = JSON.parse(fs.readFileSync("test-results/results.json", "utf8")).stats; } catch {}
  const line = { at: new Date().toISOString(), commit, exit: Number(status), secs: Number(secs),
    passed: s.expected ?? null, failed: s.unexpected ?? null, flaky: s.flaky ?? null, skipped: s.skipped ?? null };
  fs.appendFileSync(file, JSON.stringify(line) + "\n");
  console.log("e2e: " + JSON.stringify(line));
' "$HISTORY" "$commit" "$status" "$secs"

if [[ $status -ne 0 ]]; then
  keep="$STATE/runs/$(date +%Y%m%d-%H%M%S)-$commit"
  mkdir -p "$keep"
  cp -r test-results playwright-report "$keep"/ 2>/dev/null
  echo "e2e: FAILED (exit $status). Report kept in $keep (failures.md, playwright-report/)."
  if [[ "${E2E_BLOCKING:-0}" == 1 ]]; then exit "$status"; fi
  echo "e2e: non-blocking (E2E_BLOCKING=0), so the pipeline continues."
fi
exit 0
