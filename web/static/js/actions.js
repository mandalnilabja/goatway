// Button actions for credentials and API keys
// Depends on: API, Pages, Modals, Router

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
    },

    async deleteAPIKey(id) {
        if (!confirm('Delete this API key? This cannot be undone.')) return;
        try {
            await API.deleteAPIKey(id);
            Pages.apikeys();
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    },

    async rotateAPIKey(id) {
        if (!confirm('Rotate this API key? The old key will stop working immediately.')) return;
        try {
            const result = await API.rotateAPIKey(id);
            Modals.showAPIKeyCreated(result.key);
            Pages.apikeys();
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    },

    async toggleAPIKey(id, currentlyActive) {
        const action = currentlyActive ? 'disable' : 'enable';
        if (!confirm(`${action.charAt(0).toUpperCase() + action.slice(1)} this API key?`)) return;
        try {
            await API.updateAPIKey(id, { is_active: !currentlyActive });
            Pages.apikeys();
        } catch (err) {
            alert('Error: ' + (err?.message || String(err)));
        }
    }
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => Router.init());
