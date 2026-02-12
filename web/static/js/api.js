// API helper functions for Goatway admin endpoints

const API = {
    baseUrl: '/api/admin',

    async request(endpoint, options = {}) {
        const url = this.baseUrl + endpoint;
        const config = {
            headers: { 'Content-Type': 'application/json' },
            ...options
        };

        const response = await fetch(url, config);

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: response.statusText }));
            const errorMsg = typeof error.error === 'string'
                ? error.error
                : (error.message || JSON.stringify(error) || 'Request failed');
            throw new Error(errorMsg);
        }

        // Handle 204 No Content or empty responses
        if (response.status === 204 || response.headers.get('content-length') === '0') {
            return null;
        }
        return response.json();
    },

    // Credentials
    async listCredentials() {
        return this.request('/credentials');
    },

    async getCredential(id) {
        return this.request(`/credentials/${id}`);
    },

    async createCredential(data) {
        return this.request('/credentials', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateCredential(id, data) {
        return this.request(`/credentials/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    async deleteCredential(id) {
        return this.request(`/credentials/${id}`, { method: 'DELETE' });
    },

    async setDefaultCredential(id) {
        return this.request(`/credentials/${id}/default`, { method: 'POST' });
    },

    // API Keys
    async listAPIKeys() {
        return this.request('/apikeys');
    },

    async getAPIKey(id) {
        return this.request(`/apikeys/${id}`);
    },

    async createAPIKey(data) {
        return this.request('/apikeys', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateAPIKey(id, data) {
        return this.request(`/apikeys/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    async deleteAPIKey(id) {
        return this.request(`/apikeys/${id}`, { method: 'DELETE' });
    },

    async rotateAPIKey(id) {
        return this.request(`/apikeys/${id}/rotate`, { method: 'POST' });
    },

    // Usage
    async getUsageStats(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request('/usage' + (query ? `?${query}` : ''));
    },

    async getDailyUsage(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request('/usage/daily' + (query ? `?${query}` : ''));
    },

    // Logs
    async getRequestLogs(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request('/logs' + (query ? `?${query}` : ''));
    },

    async deleteRequestLogs(beforeDate) {
        return this.request(`/logs?before_date=${beforeDate}`, { method: 'DELETE' });
    },

    // System
    async getHealth() {
        return this.request('/health');
    },

    async getInfo() {
        return this.request('/info');
    }
};

// Utility functions
const Utils = {
    formatNumber(num) {
        if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
        if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
        return num.toString();
    },

    formatDate(dateStr) {
        return new Date(dateStr).toLocaleDateString();
    },

    formatDateTime(dateStr) {
        const d = new Date(dateStr);
        return d.toLocaleDateString() + ' ' + d.toLocaleTimeString();
    },

    formatDuration(ms) {
        if (ms >= 1000) return (ms / 1000).toFixed(2) + 's';
        return ms + 'ms';
    },

    getDateRange(days) {
        const end = new Date();
        const start = new Date();
        start.setDate(start.getDate() - days);
        return {
            start_date: start.toISOString().split('T')[0],
            end_date: end.toISOString().split('T')[0]
        };
    }
};
