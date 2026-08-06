if ('serviceWorker' in navigator) {
	window.addEventListener('load', () => {
		navigator.serviceWorker
			.register('/service_worker.js')
			.then((registration) => {
				console.log('Service Worker registrado con éxito:', registration.scope);
			})
			.catch((error) => {
				console.log('Falló el registro del Service Worker:', error);
			});
	});
}
