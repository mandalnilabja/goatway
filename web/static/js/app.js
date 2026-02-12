// Goatway Dashboard - Main Application
// Loads: api.js, pages.js (included before this file)

// Router with History API
const Router = {
    basePath: '/web',
    routes: {
        '/web': () => Pages.dashboard(),
        '/web/credentials': () => Pages.credentials(),
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

        // Find and execute route handler
        const handler = this.routes[path];
        if (handler) {
            handler(params);
        } else {
            document.getElementById('app').innerHTML = `
                <div>
                    <div>404</div>
                    <p>Page not found</p>
                    <a href="/web" data-link>Go Home</a>
                </div>
            `;
        }
    }
};

// Chart rendering helpers
const Charts = {
    usageChart: null,
    modelChart: null,

    renderUsageChart(canvasId, dailyData) {
        const ctx = document.getElementById(canvasId);
        if (!ctx) return;

        if (this.usageChart) this.usageChart.destroy();

        const labels = dailyData.map(d => d.date);
        const tokens = dailyData.map(d => d.total_tokens || 0);

        this.usageChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels,
                datasets: [{
                    label: 'Tokens',
                    data: tokens,
                    backgroundColor: '#2563eb',
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { display: false } },
                scales: {
                    y: { beginAtZero: true },
                    x: { grid: { display: false } }
                }
            }
        });
    },

    renderModelChart(canvasId, models) {
        const ctx = document.getElementById(canvasId);
        if (!ctx) return;

        if (this.modelChart) this.modelChart.destroy();

        const entries = Object.entries(models);
        if (!entries.length) {
            ctx.parentElement.innerHTML = '<div><p>No model data yet</p></div>';
            return;
        }

        const labels = entries.map(([model]) => model);
        const data = entries.map(([, stats]) => stats.tokens || 0);
        const colors = ['#2563eb', '#16a34a', '#ca8a04', '#dc2626', '#8b5cf6', '#ec4899'];

        this.modelChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels,
                datasets: [{
                    data,
                    backgroundColor: colors.slice(0, labels.length)
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { position: 'bottom' }
                }
            }
        });
    }
};

// Modal management
const Modals = {
    show(content) {
        const overlay = document.createElement('div');
        overlay.id = 'modal-overlay';
        overlay.innerHTML = content;
        overlay.addEventListener('click', e => {
            if (e.target === overlay) this.close();
        });
        document.body.appendChild(overlay);
    },

    close() {
        document.getElementById('modal-overlay')?.remove();
    },

    async showCredentialForm(editId = null) {
        let credential = { provider: 'openrouter', name: '', api_key: '', is_default: false };

        if (editId) {
            try {
                credential = await API.getCredential(editId);
            } catch (err) {
                alert('Error loading credential: ' + (err?.message || String(err)));
                return;
            }
        }

        this.show(`
            <div>
                <div>
                    <h3>${editId ? 'Edit' : 'Add'} Credential</h3>
                    <button onclick="Modals.close()">&times;</button>
                </div>
                <form id="credential-form">
                    <div>
                        <label>Name</label>
                        <input type="text" name="name" value="${credential.name}" required placeholder="My API Key">
                    </div>
                    <div>
                        <label>Provider</label>
                        <select name="provider" required>
                            <option value="openrouter" ${credential.provider === 'openrouter' ? 'selected' : ''}>OpenRouter</option>
                            <option value="openai" ${credential.provider === 'openai' ? 'selected' : ''}>OpenAI</option>
                            <option value="anthropic" ${credential.provider === 'anthropic' ? 'selected' : ''}>Anthropic</option>
                        </select>
                    </div>
                    <div>
                        <label>API Key</label>
                        <input type="password" name="api_key" ${editId ? '' : 'required'} placeholder="${editId ? 'Leave blank to keep current' : 'sk-...'}">
                    </div>
                    <div>
                        <label><input type="checkbox" name="is_default" ${credential.is_default ? 'checked' : ''}> Set as default</label>
                    </div>
                    <div>
                        <button type="button" onclick="Modals.close()">Cancel</button>
                        <button type="submit">${editId ? 'Update' : 'Create'}</button>
                    </div>
                </form>
            </div>
        `);

        document.getElementById('credential-form').onsubmit = async (e) => {
            e.preventDefault();
            const form = e.target;
            const data = {
                name: form.name.value,
                provider: form.provider.value,
                is_default: form.is_default.checked
            };
            if (form.api_key.value) data.api_key = form.api_key.value;

            try {
                if (editId) {
                    await API.updateCredential(editId, data);
                } else {
                    await API.createCredential(data);
                }
                this.close();
                Pages.credentials();
            } catch (err) {
                alert('Error: ' + (err?.message || String(err)));
            }
        };
    }
};

// Actions for buttons
const Actions = {
    async setDefault(id) {
        if (!confirm('Set this credential as default?')) return;
        try {
            await API.setDefaultCredential(id);
            Pages.credentials();
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    },

    async deleteCredential(id) {
        if (!confirm('Delete this credential? This cannot be undone.')) return;
        try {
            await API.deleteCredential(id);
            Pages.credentials();
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    },

    async clearOldLogs() {
        if (!confirm('Delete all logs older than 30 days?')) return;
        const date = new Date();
        date.setDate(date.getDate() - 30);
        try {
            await API.deleteRequestLogs(date.toISOString().split('T')[0]);
            alert('Old logs cleared successfully');
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    }
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => Router.init());
