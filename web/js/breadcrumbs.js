/**
 * Dynamic Header Breadcrumbs Generator
 *
 * Parses the current URL path to build dynamic, clickable breadcrumb links
 * inside the header element (#header-page-title) for intuitive parent navigation.
 */
document.addEventListener('DOMContentLoaded', () => {
	const titleContainer = document.getElementById('header-page-title');
	if (!titleContainer) return;

	const currentPath = window.location.pathname;

	// Direct route for homepage
	if (currentPath === '/' || currentPath === '') {
		titleContainer.textContent = 'Inicio';
		return;
	}

	// Friendly display names for path segments
	const routeLabels = {
		guides: 'Guías',
		manual_bicho: 'Manual del Bicho',
		news: 'Novedades',
		calculo_notas: 'Guía de Notas',
		tools: 'Herramientas',
		interactive_graph: 'Mallas Interactivas',
		calculator: 'Calculadora',
	};

	const segments = currentPath.split('/').filter(Boolean);
	let accumulatedPath = '';
	const breadcrumbLinks = [];

	segments.forEach((segment, index) => {
		accumulatedPath += `/${segment}`;
		const isLast = index === segments.length - 1;

		// Label lookup or fallback formatting (replaces underscores/hyphens with spaces)
		const label = routeLabels[segment] || segment.replace(/[-_]/g, ' ');
		const formattedLabel = label.charAt(0).toUpperCase() + label.slice(1);

		if (isLast) {
			// Current active page (Non-clickable text)
			breadcrumbLinks.push(`
<span class="font-semibold text-gray-800">${formattedLabel}</span>
`);
		} else {
			// Parent level (Clickable link for back navigation)
			breadcrumbLinks.push(` <a href="${accumulatedPath}" class="text-gray-500 hover:text-primary-600 transition-colors">
                                    ${formattedLabel} </a>`);
		}
	});

	// Render breadcrumbs separated by a '/'
	titleContainer.innerHTML = breadcrumbLinks.join('<span class="mx-2 text-gray-400 font-normal">/</span>');
});
