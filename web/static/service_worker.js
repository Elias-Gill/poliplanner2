/**
 * @file sw.js
 * @description PoliPlanner Service Worker
 */

const CACHE_NAME = 'poliplanner-cache-v2';

self.addEventListener('install', (event) => {
	self.skipWaiting();
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches.keys().then((cacheNames) => {
			return Promise.all(
				cacheNames.map((cacheName) => {
					if (cacheName !== CACHE_NAME) {
						return caches.delete(cacheName);
					}
				}),
			);
		}),
	);
	self.clients.claim();
});

// Offline error page
const ERROR_HTML = `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Sin conexión - PoliPlanner</title>
    <style>
        body { background-color: #f9fafb; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 16px; margin: 0; font-family: system-ui, -apple-system, sans-serif; }
        .card { max-width: 28rem; width: 100%; background-color: #ffffff; padding: 32px; border-radius: 2px; box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05); border: 1px solid #e5e7eb; display: flex; flex-direction: column; gap: 24px; }
        .alert-box { background-color: #fef2f2; border: 1px solid #fecaca; color: #b91c1c; padding: 12px; border-radius: 2px; font-size: 12px; font-weight: 500; }
        h1 { font-size: 20px; font-weight: 700; color: #111827; margin: 0 0 4px 0; }
        p { font-size: 12px; color: #6b7280; line-height: 1.625; margin: 0; }
        .section-group { display: flex; flex-direction: column; gap: 8px; }
        .section-title { font-size: 12px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: #9ca3af; margin-bottom: 4px; }
        .links-list { display: flex; flex-direction: column; gap: 8px; }
        .nav-link { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; border-radius: 2px; border: 1px solid #e5e7eb; background-color: #ffffff; color: #1f2937; font-size: 12px; font-weight: 500; text-decoration: none; }
        .nav-link:hover { background-color: #f9fafb; }
        .arrow { color: #2563eb; font-weight: 700; }
    </style>
</head>
<body>
    <div class="card">
        <div class="alert-box">¡Atención! No estás conectado a internet o el servidor se encuentra caído.</div>
        <div>
            <h1>PoliPlanner Offline</h1>
            <p>No se pudo establecer conexión. Ten en cuenta que podrás ingresar a las guías y otras secciones siempre y cuando las hayas visitado anteriormente en este dispositivo.</p>
        </div>
        <div class="section-group">
            <span class="section-title">Secciones disponibles</span>
            <div class="links-list">
                <a href="/dashboard" class="nav-link"><span>Dashboard</span><span class="arrow">→</span></a>
                <a href="/guides" class="nav-link"><span>Guías</span><span class="arrow">→</span></a>
                <a href="/tools/calculator" class="nav-link"><span>Calculadora</span><span class="arrow">→</span></a>
            </div>
        </div>
    </div>
</body>
</html>`;

self.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') return;

	const url = new URL(event.request.url);
	const path = url.pathname;

	const isStatic = path.startsWith('/static/');
	const isTarget = path.startsWith('/dashboard') || path.startsWith('/guides') || path.startsWith('/tools/calculator');

	// Static file: Network-First -> Cache Fallback -> 404
	if (isStatic) {
		event.respondWith(
			fetch(event.request)
				.then((networkResponse) => {
					if (networkResponse.status === 200) {
						const responseClone = networkResponse.clone();
						caches.open(CACHE_NAME).then((cache) => {
							cache.put(event.request, responseClone);
						});
					}
					return networkResponse;
				})
				.catch(() => {
					// Try cached files
					return caches.match(event.request).then((cachedResponse) => {
						if (cachedResponse) {
							return cachedResponse;
						}
						return new Response('Not Found', {
							status: 404,
							statusText: 'Not Found',
						});
					});
				}),
		);
		return;
	}

	// Network-First for the rest
	event.respondWith(
		fetch(event.request)
			.then((networkResponse) => {
				// Cache target routes
				if (isTarget && networkResponse.status === 200) {
					const responseClone = networkResponse.clone();
					caches.open(CACHE_NAME).then((cache) => {
						cache.put(event.request, responseClone);
					});
				}

				// No cache non target routes
				return networkResponse;
			})
			.catch(() => {
				// Offline
				if (isTarget) {
					// Serve cached target
					return caches.match(event.request).then((cachedResponse) => {
						if (cachedResponse) {
							return cachedResponse.text().then((html) => {
								const modifiedHtml = html.replace(
									'</body>',
									`
                                    <div id="offline-banner" style="position: fixed; bottom: 16px; right: 16px; background-color: #111827; color: #ffffff; font-size: 12px; font-family: system-ui, -apple-system, sans-serif; padding: 12px 16px; border-radius: 4px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); z-index: 99999; display: flex; align-items: center; gap: 12px; border: 1px solid #374151;">
                                        <span>⚠️ Estás viendo una versión guardada de esta página (sin conexión).</span>
                                        <button onclick="this.parentElement.remove()" style="background: transparent; border: none; color: #9ca3af; font-size: 16px; font-weight: bold; cursor: pointer; padding: 0;">×</button>
                                    </div>
                                    </body>
                                `,
								);
								return new Response(modifiedHtml, {
									status: cachedResponse.status,
									statusText: cachedResponse.statusText,
									headers: cachedResponse.headers,
								});
							});
						}
						// No cache found
						return new Response(ERROR_HTML, {
							status: 503,
							headers: { 'Content-Type': 'text/html; charset=utf-8' },
						});
					});
				}

				// Error page on non target routes
				return new Response(ERROR_HTML, {
					status: 503,
					headers: { 'Content-Type': 'text/html; charset=utf-8' },
				});
			}),
	);
});
