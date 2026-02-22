// Router with History API
// Depends on: Pages (loaded after this file)

const Router = {
    basePath: '/web',
    routes: {
        '/web': () => Pages.dashboard(),
        '/web/credentials': () => Pages.credentials(),
        '/web/apikeys': () => Pages.apikeys(),
        '/web/usage': () => Pages.usage(),
        '/web/logs': () => Pages.logs(),
        '/web/settings': () => Pages.settings()
    },

    init() {
        window.addEventListener('popstate', () => this.handleRoute());
        document.addEventListener('click', e => {
            const link = e.target.closest('[data-link], nav a');
            if (link && link.getAttribute('href')?.startsWith('/web')) {
                e.preventDefault();
                this.navigate(link.getAttribute('href'));
            }
        });
        this.handleRoute();
    },

    navigate(path) {
        history.pushState(null, '', path);
        this.handleRoute();
    },

    handleRoute() {
        const path = window.location.pathname;
        const params = Object.fromEntries(new URLSearchParams(window.location.search));

        // Update active nav link
        document.querySelectorAll('nav a').forEach(a => {
            a.classList.toggle('active', a.getAttribute('href') === path);
        });

        // Find and execute route handler
        const handler = this.routes[path];
        if (handler) {
            handler(params);
        } else {
            document.getElementById('app').innerHTML = `
                <div class="empty-state">
                    <div class="stat-value">404</div>
                    <p>Page not found</p>
                    <a href="/web" data-link class="btn btn-primary">Go Home</a>
                </div>
            `;
        }
    }
};
