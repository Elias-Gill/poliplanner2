/**
 * Active Navigation & Sidebar Link Highlighter
 *
 * Automatically highlights the active link and icon in the sidebar based on 
 * the current URL path.
 */
document.addEventListener('DOMContentLoaded', () => {
    const sidebar = document.getElementById('main-sidebar');
    if (!sidebar) return;

    const currentPath = window.location.pathname;

    sidebar.querySelectorAll('a.nav-link').forEach((link) => {
        const linkPath = new URL(link.href).pathname;

        if (currentPath === linkPath || (linkPath !== '/' && currentPath.startsWith(linkPath))) {
            link.classList.add('bg-white/10', 'text-blue-400', 'font-semibold', 'shadow-sm');
            link.classList.remove('text-gray-700');
            
            const icon = link.querySelector('svg');
            if (icon) {
                icon.classList.remove('text-gray-400');
                icon.classList.add('text-blue-400');
            }
        }
    });
});
