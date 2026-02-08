// Page rendering functions for Goatway dashboard

const Pages = {
    // Dashboard page
    async dashboard() {
        const app = document.getElementById('app');
        app.innerHTML = '<div class="loading"><div class="spinner"></div></div>';

        try {
            const [usage, logs, info] = await Promise.all([
                API.getUsageStats(Utils.getDateRange(1)),
                API.getRequestLogs({ limit: 10 }),
                API.getInfo()
            ]);

            app.innerHTML = `
                <div class="section-header">
                    <h2 class="section-title">Dashboard</h2>
                </div>

                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-value">${Utils.formatNumber(usage.total_requests || 0)}</div>
                        <div class="stat-label">Requests Today</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${Utils.formatNumber(usage.total_tokens || 0)}</div>
                        <div class="stat-label">Tokens Today</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${usage.error_count || 0}</div>
                        <div class="stat-label">Errors Today</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${info.uptime || 'N/A'}</div>
                        <div class="stat-label">Uptime</div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <h3 class="card-title">Recent Requests</h3>
                        <a href="/logs" class="btn btn-secondary btn-sm" data-link>View All</a>
                    </div>
                    <div class="table-container">
                        ${this.renderLogsTable(logs.logs || [])}
                    </div>
                </div>
            `;
        } catch (err) {
            app.innerHTML = `<div class="alert alert-danger">Error loading dashboard: ${err.message}</div>`;
        }
    },

    // Credentials page
    async credentials() {
        const app = document.getElementById('app');
        app.innerHTML = '<div class="loading"><div class="spinner"></div></div>';

        try {
            const creds = await API.listCredentials();

            app.innerHTML = `
                <div class="section-header">
                    <h2 class="section-title">API Credentials</h2>
                    <button class="btn btn-primary" onclick="Modals.showCredentialForm()">+ Add New</button>
                </div>

                <div id="credentials-list">
                    ${creds.length ? creds.map(c => this.renderCredentialCard(c)).join('') :
                      '<div class="empty-state"><div class="empty-state-icon">🔑</div><p>No credentials configured yet.</p></div>'}
                </div>
            `;
        } catch (err) {
            app.innerHTML = `<div class="alert alert-danger">Error loading credentials: ${err.message}</div>`;
        }
    },

    renderCredentialCard(cred) {
        return `
            <div class="credential-card" data-id="${cred.id}">
                <div class="credential-header">
                    <div class="credential-name">
                        ${cred.is_default ? '<span class="badge badge-default">Default</span>' : ''}
                        ${cred.name}
                    </div>
                    <div class="credential-actions">
                        ${!cred.is_default ? `<button class="btn btn-secondary btn-sm" onclick="Actions.setDefault('${cred.id}')">Set Default</button>` : ''}
                        <button class="btn btn-secondary btn-sm" onclick="Modals.showCredentialForm('${cred.id}')">Edit</button>
                        <button class="btn btn-danger btn-sm" onclick="Actions.deleteCredential('${cred.id}')">Delete</button>
                    </div>
                </div>
                <div class="credential-meta">
                    Provider: ${cred.provider} | Key: ${cred.api_key_preview || '***'} | Created: ${Utils.formatDate(cred.created_at)}
                </div>
            </div>
        `;
    },

    // Usage page
    async usage() {
        const app = document.getElementById('app');
        app.innerHTML = '<div class="loading"><div class="spinner"></div></div>';

        try {
            const range = Utils.getDateRange(7);
            const [stats, daily] = await Promise.all([
                API.getUsageStats(range),
                API.getDailyUsage(range)
            ]);

            app.innerHTML = `
                <div class="section-header">
                    <h2 class="section-title">Usage Analytics</h2>
                </div>

                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-value">${Utils.formatNumber(stats.total_requests || 0)}</div>
                        <div class="stat-label">Total Requests (7d)</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${Utils.formatNumber(stats.total_tokens || 0)}</div>
                        <div class="stat-label">Total Tokens (7d)</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${Utils.formatNumber(stats.prompt_tokens || 0)}</div>
                        <div class="stat-label">Prompt Tokens</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${Utils.formatNumber(stats.completion_tokens || 0)}</div>
                        <div class="stat-label">Completion Tokens</div>
                    </div>
                </div>

                <div class="grid-2">
                    <div class="card">
                        <h3 class="card-title">Token Usage Over Time</h3>
                        <div class="chart-container">
                            <canvas id="usage-chart"></canvas>
                        </div>
                    </div>
                    <div class="card">
                        <h3 class="card-title">Model Breakdown</h3>
                        <div class="chart-container">
                            <canvas id="model-chart"></canvas>
                        </div>
                    </div>
                </div>
            `;

            Charts.renderUsageChart('usage-chart', daily);
            Charts.renderModelChart('model-chart', stats.models || {});
        } catch (err) {
            app.innerHTML = `<div class="alert alert-danger">Error loading usage: ${err.message}</div>`;
        }
    },

    // Logs page
    async logs(params = {}) {
        const app = document.getElementById('app');
        const limit = params.limit || 50;
        const offset = params.offset || 0;

        app.innerHTML = '<div class="loading"><div class="spinner"></div></div>';

        try {
            const data = await API.getRequestLogs({ limit, offset, ...params });

            app.innerHTML = `
                <div class="section-header">
                    <h2 class="section-title">Request Logs</h2>
                </div>

                <div class="card">
                    <div class="table-container">
                        ${this.renderLogsTable(data.logs || [], true)}
                    </div>
                    ${this.renderPagination(data.total, limit, offset)}
                </div>
            `;
        } catch (err) {
            app.innerHTML = `<div class="alert alert-danger">Error loading logs: ${err.message}</div>`;
        }
    },

    renderLogsTable(logs, showMore = false) {
        if (!logs.length) {
            return '<div class="empty-state"><p>No request logs yet.</p></div>';
        }

        return `
            <table>
                <thead>
                    <tr>
                        <th>Time</th>
                        <th>Model</th>
                        <th>Tokens</th>
                        <th>Duration</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    ${logs.map(log => `
                        <tr>
                            <td>${Utils.formatDateTime(log.created_at)}</td>
                            <td>${log.model}</td>
                            <td>${log.total_tokens || '-'}</td>
                            <td>${log.duration_ms ? Utils.formatDuration(log.duration_ms) : '-'}</td>
                            <td>
                                <span class="badge ${log.status_code === 200 ? 'badge-success' : 'badge-danger'}">
                                    ${log.status_code}
                                </span>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
    },

    renderPagination(total, limit, offset) {
        const currentPage = Math.floor(offset / limit) + 1;
        const totalPages = Math.ceil(total / limit);

        return `
            <div class="pagination">
                <button class="pagination-btn" onclick="Router.navigate('/logs?offset=${Math.max(0, offset - limit)}')" ${offset === 0 ? 'disabled' : ''}>Previous</button>
                <span class="pagination-info">Page ${currentPage} of ${totalPages}</span>
                <button class="pagination-btn" onclick="Router.navigate('/logs?offset=${offset + limit}')" ${offset + limit >= total ? 'disabled' : ''}>Next</button>
            </div>
        `;
    },

    // Settings page
    async settings() {
        const app = document.getElementById('app');
        app.innerHTML = '<div class="loading"><div class="spinner"></div></div>';

        try {
            const info = await API.getInfo();

            app.innerHTML = `
                <div class="section-header">
                    <h2 class="section-title">Settings</h2>
                </div>

                <div class="card" style="margin-bottom: 1rem;">
                    <h3 class="card-title">System Information</h3>
                    <table>
                        <tr><td><strong>Version</strong></td><td>${info.version || 'dev'}</td></tr>
                        <tr><td><strong>Uptime</strong></td><td>${info.uptime || 'N/A'}</td></tr>
                        <tr><td><strong>Data Directory</strong></td><td>${info.data_dir || 'N/A'}</td></tr>
                    </table>
                </div>

                <div class="card">
                    <h3 class="card-title">Danger Zone</h3>
                    <p style="margin-bottom: 1rem; color: var(--text-muted);">Clear old request logs to free up disk space.</p>
                    <button class="btn btn-danger" onclick="Actions.clearOldLogs()">Clear Logs Older Than 30 Days</button>
                </div>
            `;
        } catch (err) {
            app.innerHTML = `<div class="alert alert-danger">Error loading settings: ${err.message}</div>`;
        }
    }
};
