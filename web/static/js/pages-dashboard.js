// Page rendering - Dashboard and shared helpers
// Depends on: API, Utils

const Pages = {
    // Dashboard page
    async dashboard() {
        const app = document.getElementById('app');
        app.innerHTML = '<div class="loading"><div class="spinner"></div>Loading...</div>';

        try {
            const [usage, logs, info] = await Promise.all([
                API.getUsageStats(Utils.getDateRange(1)),
                API.getRequestLogs({ limit: 10 }),
                API.getInfo()
            ]);

            app.innerHTML = `
                <div class="page-header"><h2>Dashboard</h2></div>

                <div class="stats-grid">
                    <div class="card stat-card">
                        <div class="stat-value">${Utils.formatNumber(usage.total_requests || 0)}</div>
                        <div class="stat-label">Requests Today</div>
                    </div>
                    <div class="card stat-card">
                        <div class="stat-value">${Utils.formatNumber(usage.total_tokens || 0)}</div>
                        <div class="stat-label">Tokens Today</div>
                    </div>
                    <div class="card stat-card">
                        <div class="stat-value">${usage.error_count || 0}</div>
                        <div class="stat-label">Errors Today</div>
                    </div>
                    <div class="card stat-card">
                        <div class="stat-value">${info.uptime || 'N/A'}</div>
                        <div class="stat-label">Uptime</div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <h3>Recent Requests</h3>
                        <a href="/web/logs" data-link>View All</a>
                    </div>
                    <div class="table-container">
                        ${this.renderLogsTable(logs.logs || [])}
                    </div>
                </div>
            `;
        } catch (err) {
            app.innerHTML = `<div class="error">Error loading dashboard: ${err?.message || err}</div>`;
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
                    ${logs.map(log => {
                        const statusClass = log.status_code >= 400 ? 'badge-danger' :
                                           log.status_code >= 300 ? 'badge-warning' : 'badge-success';
                        return `
                            <tr>
                                <td>${Utils.formatDateTime(log.created_at)}</td>
                                <td>${log.model}</td>
                                <td>${log.total_tokens || '-'}</td>
                                <td>${log.duration_ms ? Utils.formatDuration(log.duration_ms) : '-'}</td>
                                <td><span class="badge ${statusClass}">${log.status_code}</span></td>
                            </tr>
                        `;
                    }).join('')}
                </tbody>
            </table>
        `;
    },

    renderPagination(total, limit, offset) {
        const currentPage = Math.floor(offset / limit) + 1;
        const totalPages = Math.ceil(total / limit);

        return `
            <div class="pagination">
                <button onclick="Router.navigate('/web/logs?offset=${Math.max(0, offset - limit)}')" ${offset === 0 ? 'disabled' : ''}>Previous</button>
                <span>Page ${currentPage} of ${totalPages}</span>
                <button onclick="Router.navigate('/web/logs?offset=${offset + limit}')" ${offset + limit >= total ? 'disabled' : ''}>Next</button>
            </div>
        `;
    }
};
