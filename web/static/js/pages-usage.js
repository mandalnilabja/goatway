// Usage analytics page
// Extends: Pages (must be loaded after pages-dashboard.js)
// Depends on: API, Utils, Charts

Pages.usage = async function() {
    const app = document.getElementById('app');
    app.innerHTML = '<div class="loading"><div class="spinner"></div>Loading...</div>';

    try {
        const range = Utils.getDateRange(7);
        const [stats, daily] = await Promise.all([
            API.getUsageStats(range),
            API.getDailyUsage(range)
        ]);

        app.innerHTML = `
            <div class="page-header"><h2>Usage Analytics</h2></div>

            <div class="stats-grid">
                <div class="card stat-card">
                    <div class="stat-value">${Utils.formatNumber(stats.total_requests || 0)}</div>
                    <div class="stat-label">Total Requests (7d)</div>
                </div>
                <div class="card stat-card">
                    <div class="stat-value">${Utils.formatNumber(stats.total_tokens || 0)}</div>
                    <div class="stat-label">Total Tokens (7d)</div>
                </div>
                <div class="card stat-card">
                    <div class="stat-value">${Utils.formatNumber(stats.prompt_tokens || 0)}</div>
                    <div class="stat-label">Prompt Tokens</div>
                </div>
                <div class="card stat-card">
                    <div class="stat-value">${Utils.formatNumber(stats.completion_tokens || 0)}</div>
                    <div class="stat-label">Completion Tokens</div>
                </div>
            </div>

            <div class="charts-grid">
                <div class="card">
                    <h3>Token Usage Over Time</h3>
                    <div class="chart-container">
                        <canvas id="usage-chart"></canvas>
                    </div>
                </div>
                <div class="card">
                    <h3>Model Breakdown</h3>
                    <div class="chart-container">
                        <canvas id="model-chart"></canvas>
                    </div>
                </div>
            </div>
        `;

        Charts.renderUsageChart('usage-chart', daily.daily_usage || []);
        Charts.renderModelChart('model-chart', stats.models || {});
    } catch (err) {
        app.innerHTML = `<div class="error">Error loading usage: ${err?.message || err}</div>`;
    }
};
