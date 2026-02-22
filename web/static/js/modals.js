// Modal management - Base functionality
// Extended by: modals-credential.js, modals-apikey.js

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
    }
};
