import fs from 'node:fs';
import http from 'node:http';
import type net from 'node:net';
import path from 'node:path';

/**
 * The one origin the browser talks to. It serves the frontend (the cached production build, or a
 * vite dev server with E2E_FRONTEND=dev) and forwards /api/* to the backend named by the
 * `e2e_backend` cookie, so every worker shares one frontend but gets its own backend and database.
 * Mirrors frontend/nginx.conf: SPA fallback, real 404 for missing /assets/.
 */
export const BACKEND_COOKIE = 'e2e_backend';

const TYPES: Record<string, string> = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript',
  '.mjs': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.webmanifest': 'application/manifest+json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.ico': 'image/x-icon',
  '.woff2': 'font/woff2',
  '.txt': 'text/plain',
};

function backendPort(req: http.IncomingMessage): number | undefined {
  const m = (req.headers.cookie ?? '').match(new RegExp(`(?:^|;\\s*)${BACKEND_COOKIE}=(\\d+)`));
  const port = m ? Number(m[1]) : NaN;
  return port > 1024 && port < 65536 ? port : undefined;
}

function proxy(
  req: http.IncomingMessage,
  res: http.ServerResponse,
  port: number,
  pathname: string,
): void {
  const upstream = http.request(
    { host: '127.0.0.1', port, method: req.method, path: pathname, headers: { ...req.headers, host: `127.0.0.1:${port}` } },
    (up) => {
      res.writeHead(up.statusCode ?? 502, up.headers);
      up.pipe(res);
      up.on('error', () => res.destroy());
    },
  );
  // A dropped upstream connection (an e2e "drop" hold) must reach the browser as a network
  // error, not as a 502 page: destroy the client socket instead of answering.
  upstream.on('error', () => res.destroy());
  req.pipe(upstream);
}

function serveStatic(dist: string, pathname: string, res: http.ServerResponse): void {
  const rel = decodeURIComponent(pathname).replace(/^\/+/, '');
  let file = path.join(dist, rel);
  if (!file.startsWith(dist)) {
    res.writeHead(400).end();
    return;
  }
  const isFile = fs.existsSync(file) && fs.statSync(file).isFile();
  if (!isFile) {
    if (pathname.startsWith('/assets/')) {
      res.writeHead(404).end();
      return;
    }
    file = path.join(dist, 'index.html');
  }
  const type = TYPES[path.extname(file)] ?? 'application/octet-stream';
  const cache = pathname.startsWith('/assets/') ? 'public, max-age=31536000, immutable' : 'no-cache';
  res.writeHead(200, { 'content-type': type, 'cache-control': cache });
  fs.createReadStream(file).pipe(res);
}

export interface FrontDoor {
  url: string;
  close(): Promise<void>;
}

/** `dist` serves a production build; `devServer` (a vite URL) proxies everything else to vite. */
export function startFrontDoor(opts: { dist?: string; devServer?: string }): Promise<FrontDoor> {
  const dev = opts.devServer ? new URL(opts.devServer) : undefined;
  const server = http.createServer((req, res) => {
    const url = new URL(req.url ?? '/', 'http://x');
    if (url.pathname === '/api' || url.pathname.startsWith('/api/')) {
      const port = backendPort(req);
      if (!port) {
        res.writeHead(502, { 'content-type': 'text/plain' }).end(`front door: no ${BACKEND_COOKIE} cookie`);
        return;
      }
      proxy(req, res, port, (url.pathname.slice(4) || '/') + url.search);
      return;
    }
    if (dev) {
      proxy(req, res, Number(dev.port), url.pathname + url.search);
      return;
    }
    serveStatic(opts.dist!, url.pathname, res);
  });
  // vite's HMR websocket, in dev mode.
  server.on('upgrade', (req, socket: net.Socket, head) => {
    if (!dev) return socket.destroy();
    const up = http.request({
      host: '127.0.0.1',
      port: Number(dev.port),
      path: req.url,
      method: req.method,
      headers: req.headers,
    });
    up.on('upgrade', (upRes, upSocket, upHead) => {
      socket.write(
        `HTTP/1.1 101 Switching Protocols\r\n` +
          Object.entries(upRes.headers)
            .map(([k, v]) => `${k}: ${v}`)
            .join('\r\n') +
          '\r\n\r\n',
      );
      if (upHead.length) socket.write(upHead);
      if (head.length) upSocket.write(head);
      upSocket.pipe(socket).pipe(upSocket);
    });
    up.on('error', () => socket.destroy());
    up.end();
  });
  server.keepAliveTimeout = 30_000;
  return new Promise((resolve) => {
    server.listen(0, '127.0.0.1', () => {
      const { port } = server.address() as net.AddressInfo;
      resolve({
        url: `http://127.0.0.1:${port}`,
        close: () =>
          new Promise((r) => {
            server.closeAllConnections();
            server.close(() => r());
          }),
      });
    });
  });
}
