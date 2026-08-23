(function () {
	const root = document.documentElement;
	root.classList.toggle(
		'dark',
		localStorage.mode === 'dark' ||
			(!('mode' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches),
	);
})();
