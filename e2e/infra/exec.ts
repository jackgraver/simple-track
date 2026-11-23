import { spawn, type SpawnOptions } from 'node:child_process';

/** Run a command to completion; rejects with its output on a non-zero exit. */
export function run(cmd: string, args: string[], opts: SpawnOptions = {}): Promise<string> {
  return new Promise((resolve, reject) => {
    const child = spawn(cmd, args, { ...opts, stdio: ['ignore', 'pipe', 'pipe'] });
    let out = '';
    child.stdout!.on('data', (d) => (out += d));
    child.stderr!.on('data', (d) => (out += d));
    child.on('error', reject);
    child.on('close', (code) => {
      if (code === 0) resolve(out);
      else reject(new Error(`${cmd} ${args.join(' ')} exited ${code}\n${out.slice(-4000)}`));
    });
  });
}

export const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));
