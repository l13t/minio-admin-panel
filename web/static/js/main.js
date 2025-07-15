// Common JavaScript functionality for MinIO Admin Panel

// Utility functions
const Utils = {
    // Format bytes to human readable format
    formatBytes(bytes, decimals = 2) {
        if (bytes === 0) return '0 Bytes';

        const k = 1024;
        const dm = decimals < 0 ? 0 : decimals;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];

        const i = Math.floor(Math.log(bytes) / Math.log(k));

        return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
    },

    // Format date to local string
    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString();
    },

    // Show loading spinner
    showLoading(element) {
        const spinner = document.createElement('div');
        spinner.className = 'loading';
        spinner.id = 'loading-spinner';
        element.appendChild(spinner);
    },

    // Hide loading spinner
    hideLoading() {
        const spinner = document.getElementById('loading-spinner');
        if (spinner) {
            spinner.remove();
        }
    },

    // Show toast notification
    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `alert alert-${type} alert-dismissible fade show position-fixed`;
        toast.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
        toast.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;

        document.body.appendChild(toast);

        // Auto-hide after 5 seconds
        setTimeout(() => {
            if (toast.parentNode) {
                toast.remove();
            }
        }, 5000);
    },

    // Copy text to clipboard
    async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            this.showToast(this.t('ui.copied_to_clipboard'), 'success');
        } catch (err) {
            console.error('Failed to copy: ', err);
            this.showToast('Failed to copy to clipboard', 'danger');
        }
    }
};

// API helper functions
const API = {
    // Base API request
    async request(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
        };

        const config = {
            ...defaultOptions,
            ...options,
            headers: {
                ...defaultOptions.headers,
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, config);

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    },

    // GET request
    async get(url) {
        return this.request(url, { method: 'GET' });
    },

    // POST request
    async post(url, data) {
        return this.request(url, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    // PUT request
    async put(url, data) {
        return this.request(url, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    // DELETE request
    async delete(url) {
        return this.request(url, { method: 'DELETE' });
    }
};

// Form validation helpers
const Validation = {
    // Validate bucket name
    validateBucketName(name) {
        const bucketNameRegex = /^[a-z0-9][a-z0-9\-]{1,61}[a-z0-9]$/;
        return bucketNameRegex.test(name);
    },

    // Validate access key
    validateAccessKey(accessKey) {
        return accessKey.length >= 3 && accessKey.length <= 20 && /^[a-zA-Z0-9]+$/.test(accessKey);
    },

    // Validate secret key
    validateSecretKey(secretKey) {
        return secretKey.length >= 8;
    },

    // Show form validation error
    showFieldError(fieldId, message) {
        const field = document.getElementById(fieldId);
        const errorDiv = document.createElement('div');
        errorDiv.className = 'invalid-feedback';
        errorDiv.textContent = message;

        field.classList.add('is-invalid');

        // Remove existing error message
        const existingError = field.parentNode.querySelector('.invalid-feedback');
        if (existingError) {
            existingError.remove();
        }

        field.parentNode.appendChild(errorDiv);
    },

    // Clear form validation errors
    clearFieldErrors(formId) {
        const form = document.getElementById(formId);
        const invalidFields = form.querySelectorAll('.is-invalid');
        const errorMessages = form.querySelectorAll('.invalid-feedback');

        invalidFields.forEach(field => field.classList.remove('is-invalid'));
        errorMessages.forEach(error => error.remove());
    }
};

// Initialize common functionality when DOM is loaded
document.addEventListener('DOMContentLoaded', function () {
    // Add click handlers for copy buttons
    document.querySelectorAll('.copy-btn').forEach(btn => {
        btn.addEventListener('click', function () {
            const text = this.dataset.copy;
            Utils.copyToClipboard(text);
        });
    });

    // Add confirmation dialogs for delete buttons
    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', function (e) {
            const itemName = this.dataset.name;
            if (!confirm(`Are you sure you want to delete "${itemName}"?`)) {
                e.preventDefault();
            }
        });
    });

    // Auto-hide alerts after 5 seconds
    document.querySelectorAll('.alert:not(.alert-permanent)').forEach(alert => {
        setTimeout(() => {
            if (alert.parentNode) {
                alert.remove();
            }
        }, 5000);
    });
});

// Sidebar toggle for mobile
function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    sidebar.classList.toggle('show');
}

// Export for use in other scripts
window.Utils = Utils;
window.API = API;
window.Validation = Validation;
