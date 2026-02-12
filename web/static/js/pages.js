// Page rendering functions for Goatway dashboard

const Pages = {
    // Dashboard page
    async dashboard() {
        const app = document.getElementById('app');
        app.innerHTML = '<div>Loading...</div>';

        try {
            const [usage, logs, info] = await Promise.all([
                API.getUsageStats(Utils.getDateRange(1)),
                API.getRequestLogs({ limit: 10 }),
                API.getInfo()
            ]);

            app.innerHTML = `
                <div>
                    <h2>Dashboard</h2>
                </div>

                <div>
                    <div>
                        <div>${Utils.formatNumber(usage.total_requests || 0)}</div>
                        <div>Requests Today</div>
                    </div>
                    <div>
                        <div>${Utils.formatNumber(usage.total_tokens || 0)}</div>
                        <div>Tokens Today</div>
                    </div>
                    <div>
                        <div>${usage.error_count || 0}</div>
                        <div>Errors Today</div>
                    </div>
                    <div>
                        <div>${info.uptime || 'N/A'}</div>
                        <div>Uptime</div>
                    </div>
                </div>

                <div>
                    <div>
                        <h3>Recent Requests</h3>
                        <a href="/web/logs" data-link>View All</a>
                    </div>
                    <div>
                        ${this.renderLogsTable(logs.logs || [])}
                    </div>
                </div>
            `;
        } catch (err) {
            const msg = err?.message || String(err);
            app.innerHTML = `<div>Error loading dashboard: ${msg}</div>`;
        }
    },

    // Credentials page
    async credentials() {
        const app = document.getElementById('app');
        app.innerHTML = '<div>Loading...</div>';

        try {
            const data = await API.listCredentials();
            const creds = data.credentials || [];

            app.innerHTML = `
                <div>
                    <h2>API Credentials</h2>
                    <button onclick="Modals.showCredentialForm()">+ Add New</button>
                </div>

                <div id="credentials-list">
                    ${creds.length ? creds.map(c => this.renderCredentialCard(c)).join('') :
                      '<div><p>No credentials configured yet.</p></div>'}
                </div>
            `;
        } catch (err) {
            const msg = err?.message || String(err);
            app.innerHTML = `<div>Error loading credentials: ${msg}</div>`;
        }
    },

    renderCredentialCard(cred) {
        return `
            <div data-id="${cred.id}">
                <div>
                    <div>
                        ${cred.is_default ? '<span>[Default]</span>' : ''}
                        ${cred.name}
                    </div>
                    <div>
                        ${!cred.is_default ? `<button onclick="Actions.setDefault('${cred.id}')">Set Default</button>` : ''}
                        <button onclick="Modals.showCredentialForm('${cred.id}')">Edit</button>
                        <button onclick="Actions.deleteCredential('${cred.id}')">Delete</button>
                    </div>
                </div>
                <div>
                    Provider: ${cred.provider} | Key: ${cred.api_key_preview || '***'} | Created: ${Utils.formatDate(cred.created_at)}
                </div>
            </div>
        `;
    },

    // API Keys page
    async apikeys() {
        const app = document.getElementById('app');
        app.innerHTML = '<div>Loading...</div>';

        try {
            const data = await API.listAPIKeys();
            const keys = data.data || [];

            app.innerHTML = `
                <div>
                    <h2>API Keys</h2>
                    <button onclick="Modals.showAPIKeyForm()">+ Create Key</button>
                </div>

                <p>API keys allow applications to access the proxy endpoints. Use these keys with the OpenAI SDK.</p>

                <div id="apikeys-list">
                    ${keys.length ? keys.map(k => this.renderAPIKeyCard(k)).join('') :
                      '<div><p>No API keys created yet.</p></div>'}
                </div>
            `;
        } catch (err) {
            const msg = err?.message || String(err);
            app.innerHTML = `<div>Error loading API keys: ${msg}</div>`;
        }
    },

    renderAPIKeyCard(key) {
        const statusLabel = key.is_active ? '[Active]' : '[Inactive]';
        const scopes = (key.scopes || []).join(', ');
        const rateLimit = key.rate_limit ? `${key.rate_limit}/min` : 'Unlimited';
        const expiresAt = key.expires_at ? Utils.formatDate(key.expires_at) : 'Never';

        return `
            <div data-id="${key.id}">
                <div>
                    <div>
                        <span>${statusLabel}</span>
                        ${key.name}
                    </div>
                    <div>
                        <button onclick="Actions.toggleAPIKey('${key.id}', ${key.is_active})">${key.is_active ? 'Disable' : 'Enable'}</button>
                        <button onclick="Actions.rotateAPIKey('${key.id}')">Rotate</button>
                        <button onclick="Modals.showAPIKeyForm('${key.id}')">Edit</button>
                        <button onclick="Actions.deleteAPIKey('${key.id}')">Delete</button>
                    </div>
                </div>
                <div>
                    Key: ${key.key_prefix}... | Scopes: ${scopes} | Rate: ${rateLimit} | Expires: ${expiresAt} | Created: ${Utils.formatDate(key.created_at)}
                </div>
            </div>
        `;
    },

    // Usage page
    async usage() {
        const app = document.getElementById('app');
        app.innerHTML = '<div>Loading...</div>';

        try {
            const range = Utils.getDateRange(7);
            const [stats, daily] = await Promise.all([
                API.getUsageStats(range),
                API.getDailyUsage(range)
            ]);

            app.innerHTML = `
                <div>
                    <h2>Usage Analytics</h2>
                </div>

                <div>
                    <div>
                        <div>${Utils.formatNumber(stats.total_requests || 0)}</div>
                        <div>Total Requests (7d)</div>
                    </div>
                    <div>
                        <div>${Utils.formatNumber(stats.total_tokens || 0)}</div>
                        <div>Total Tokens (7d)</div>
                    </div>
                    <div>
                        <div>${Utils.formatNumber(stats.prompt_tokens || 0)}</div>
                        <div>Prompt Tokens</div>
                    </div>
                    <div>
                        <div>${Utils.formatNumber(stats.completion_tokens || 0)}</div>
                        <div>Completion Tokens</div>
                    </div>
                </div>

                <div>
                    <div>
                        <h3>Token Usage Over Time</h3>
                        <div>
                            <canvas id="usage-chart"></canvas>
                        </div>
                    </div>
                    <div>
                        <h3>Model Breakdown</h3>
                        <div>
                            <canvas id="model-chart"></canvas>
                        </div>
                    </div>
                </div>
            `;

            Charts.renderUsageChart('usage-chart', daily);
            Charts.renderModelChart('model-chart', stats.models || {});
        } catch (err) {
            const msg = err?.message || String(err);
            app.innerHTML = `<div>Error loading usage: ${msg}</div>`;
        }
    },

    // Logs page
    async logs(params = {}) {
        const app = document.getElementById('app');
        const limit = params.limit || 50;
        const offset = params.offset || 0;

        app.innerHTML = '<div>Loading...</div>';

        try {
            const data = await API.getRequestLogs({ limit, offset, ...params });

            app.innerHTML = `
                <div>
                    <h2>Request Logs</h2>
                </div>

                <div>
                    <div>
                        ${this.renderLogsTable(data.logs || [], true)}
                    </div>
                    ${this.renderPagination(data.total, limit, offset)}
                </div>
            `;
        } catch (err) {
            const msg = err?.message || String(err);
            app.innerHTML = `<div>Error loading logs: ${msg}</div>`;
        }
    },

    renderLogsTable(logs, showMore = false) {
        if (!logs.length) {
            return '<div><p>No request logs yet.</p></div>';
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
                                <span>${log.status_code}</span>
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
            <div>
                <button onclick="Router.navigate('/web/logs?offset=${Math.max(0, offset - limit)}')" ${offset === 0 ? 'disabled' : ''}>Previous</button>
                <span>Page ${currentPage} of ${totalPages}</span>
                <button onclick="Router.navigate('/web/logs?offset=${offset + limit}')" ${offset + limit >= total ? 'disabled' : ''}>Next</button>
            </div>
        `;
    },

    // Settings page
    async settings() {
        const app = document.getElementById('app');
        app.innerHTML = '<div>Loading...</div>';

        try {
            const info = await API.getInfo();

            app.innerHTML = `
                <div>
                    <h2>Settings</h2>
                </div>

                <div>
                    <h3>System Information</h3>
                    <table>
                        <tr><td><strong>Version</strong></td><td>${info.version || 'dev'}</td></tr>
                        <tr><td><strong>Uptime</strong></td><td>${info.uptime || 'N/A'}</td></tr>
                        <tr><td><strong>Data Directory</strong></td><td>${info.data_dir || 'N/A'}</td></tr>
                    </table>
                </div>

                <div>
                    <h3>Danger Zone</h3>
                    <p>Clear old request logs to free up disk space.</p>
                    <button onclick="Actions.clearOldLogs()">Clear Logs Older Than 30 Days</button>
                </div>
            `;
        } catch (err) {
            const msg = err?.message || String(err);
            app.innerHTML = `<div>Error loading settings: ${msg}</div>`;
        }
    }
};
