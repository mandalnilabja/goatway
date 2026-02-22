// Chart rendering helpers
// Depends on: Chart.js (loaded via CDN)

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
            ctx.parentElement.innerHTML = '<div class="empty-state"><p>No model data yet</p></div>';
            return;
        }

        const labels = entries.map(([model]) => model);
        const data = entries.map(([, stats]) => stats.tokens || 0);
        const colors = ['#2563eb', '#16a34a', '#ca8a04', '#dc2626', '#8b5cf6', '#ec4899'];

        this.modelChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels,
                datasets: [{ data, backgroundColor: colors.slice(0, labels.length) }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { position: 'bottom' } }
            }
        });
    }
};
