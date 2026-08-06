document.querySelectorAll('nav a.nav-link').forEach((link) => {
	const currentPath = window.location.pathname;
	const linkPath = new URL(link.href).pathname;
	if (currentPath === linkPath || (linkPath !== '/' && currentPath.startsWith(linkPath))) {
		link.classList.add('bg-white/10', 'text-blue-400', 'font-semibold', 'shadow-sm');
		link.classList.remove('text-gray-700');
		const icon = link.querySelector('svg');
		if (icon) {
			icon.classList.remove('text-gray-400');
			icon.classList.add('text-blue-400');
		}

		const titleEl = document.getElementById('header-page-title');
		const titleText = link.querySelector('span');
		if (titleEl && titleText) {
			titleEl.textContent = titleText.textContent;
		}
	}
});
