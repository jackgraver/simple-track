import fs from 'node:fs';
import path from 'node:path';
import { SourceMapConsumer } from 'source-map-js';
import { DIST_DIR, ROOT } from './paths';

/**
 * Rewrites production-bundle stack frames ("http://127.0.0.1:1234/assets/logging-Dfq.js:1:11028")
 * to source locations ("frontend/src/pages/gym/logging/index.vue:35:15") using the build's .map
 * files, so a console error or page error in a failure report points at code an agent can open.
 * Frames it can't map are left as they are.
 */
const consumers = new Map<string, SourceMapConsumer | null>();

function consumer(asset: string): SourceMapConsumer | null {
  if (!consumers.has(asset)) {
    const map = path.join(DIST_DIR, 'assets', asset + '.map');
    consumers.set(asset, fs.existsSync(map) ? new SourceMapConsumer(JSON.parse(fs.readFileSync(map, 'utf8'))) : null);
  }
  return consumers.get(asset)!;
}

export function mapStack(text: string): string {
  return text.replace(/https?:\/\/[^\s/]+\/assets\/([^\s:)]+\.js):(\d+):(\d+)/g, (frame, asset, line, col) => {
    const c = consumer(asset);
    const pos = c?.originalPositionFor({ line: Number(line), column: Number(col) - 1 });
    if (!pos?.source) return frame;
    const src = path.relative(ROOT, path.resolve(DIST_DIR, 'assets', pos.source));
    return `${src}:${pos.line}:${(pos.column ?? 0) + 1}`;
  });
}
