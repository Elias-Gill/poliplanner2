document.addEventListener('DOMContentLoaded', () => {
	const root = document.documentElement;

	const isDark =
		localStorage.getItem('mode') === 'dark' ||
		(!localStorage.getItem('mode') && window.matchMedia('(prefers-color-scheme: dark)').matches);

	root.classList.toggle('dark', isDark);

	const toggleTheme = () => {
		const isDarkNow = root.classList.toggle('dark');
		localStorage.setItem('mode', isDarkNow ? 'dark' : 'light');
	};

	document.querySelectorAll('.theme-toggle').forEach((btn) => {
		btn.addEventListener('click', toggleTheme);
	});
});
