// Fixed-port LAN proxy for testing on a phone: http://<host>.local:$DEV_PHONE_PORT always
// forwards to whichever checkout last ran `dev phone`, so one home-screen PWA (one origin)
// works for every worktree. Started and stopped by dev/tasks/phone.
//
// The target is re-read from the state file on every request, so switching slots needs no
// restart. WebSocket upgrades (Vite HMR) are piped through as raw TCP.
import { readFileSync } from 'node:fs';
import http from 'node:http';
import net from 'node:net';

const port = Number(process.env.DEV_PHONE_PORT);
const targetFile = process.env.DEV_PHONE_TARGET_FILE;
if (!port || !targetFile) {
  console.error('phone-proxy: DEV_PHONE_PORT and DEV_PHONE_TARGET_FILE are required');
  process.exit(1);
}

// "<slot> <frontend port> <checkout path>"
function target() {
  try {
    const [slot, p, ...path] = readFileSync(targetFile, 'utf8').trim().split(' ');
    return { slot, port: Number(p), path: path.join(' ') };
  } catch {
    return null;
  }
}

function unavailable(res, msg) {
  res.writeHead(502, { 'content-type': 'text/plain; charset=utf-8', 'cache-control': 'no-store' });
  res.end(`simple-track dev phone proxy\n\n${msg}\n\nRun \`./bin/dev phone\` in the checkout you want to see.\n`);
}

const server = http.createServer((req, res) => {
  const t = target();
  if (!t) return unavailable(res, 'No checkout is selected.');
  const upstream = http.request(
    { host: '127.0.0.1', port: t.port, method: req.method, path: req.url, headers: req.headers },
    (up) => {
      res.writeHead(up.statusCode ?? 502, { ...up.headers, 'x-dev-slot': t.slot });
      up.pipe(res);
    },
  );
  upstream.on('error', () =>
    unavailable(res, `Slot ${t.slot} (${t.path}) isn't answering on port ${t.port}. Is its stack running?`),
  );
  req.pipe(upstream);
});

server.on('upgrade', (req, socket, head) => {
  const t = target();
  if (!t) return socket.destroy();
  const upstream = net.connect(t.port, '127.0.0.1', () => {
    const lines = [`${req.method} ${req.url} HTTP/${req.httpVersion}`];
    for (let i = 0; i < req.rawHeaders.length; i += 2) lines.push(`${req.rawHeaders[i]}: ${req.rawHeaders[i + 1]}`);
    upstream.write(lines.join('\r\n') + '\r\n\r\n');
    if (head.length) upstream.write(head);
    upstream.pipe(socket);
    socket.pipe(upstream);
  });
  upstream.on('error', () => socket.destroy());
  socket.on('error', () => upstream.destroy());
});

// No host: listens on :: (IPv4 too), since mDNS may hand the phone an IPv6 address.
server.listen(port, () => console.log(`phone-proxy: listening on :${port}`));
