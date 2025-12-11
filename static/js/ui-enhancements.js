// UI/UX Enhancements - dmIntegroff

class UIEnhancements {
    constructor() {
        this.notifications = [];
        this.init();
    }

    init() {
        this.setupFormValidation();
        this.setupLoadingStates();
        this.setupTooltips();
        this.setupKeyboardShortcuts();
        this.setupAutoSave();
        this.setupConfirmDialogs();
    }

    // Enhanced Notifications
    showNotification(message, type = 'info', options = {}) {
        const {
            title = '',
            duration = 5000,
            persistent = false,
            actions = []
        } = options;

        const notification = document.createElement('div');
        notification.className = `notification-enhanced ${type}`;
        
        const id = 'notification-' + Date.now();
        notification.id = id;

        const iconMap = {
            success: 'check-circle',
            error: 'alert-circle',
            warning: 'alert-triangle',
            info: 'info'
        };

        notification.innerHTML = `
            <div class="notification-header">
                <i data-lucide="${iconMap[type] || 'info'}"></i>
                ${title ? `<span class="notification-title">${title}</span>` : ''}
                <button class="notification-close" onclick="uiEnhancements.closeNotification('${id}')">
                    <i data-lucide="x"></i>
                </button>
            </div>
            <div class="notification-body">
                ${message}
                ${actions.length > 0 ? `
                    <div style="margin-top: 0.75rem; display: flex; gap: 0.5rem;">
                        ${actions.map(action => `
                            <button class="btn btn-sm ${action.style || 'btn-secondary'}" onclick="${action.onclick}">
                                ${action.label}
                            </button>
                        `).join('')}
                    </div>
                ` : ''}
            </div>
        `;

        document.body.appendChild(notification);
        this.notifications.push(id);

        // Initialize Lucide icons
        if (window.lucide) {
            window.lucide.createIcons();
        }

        // Auto-close if not persistent
        if (!persistent && duration > 0) {
            setTimeout(() => {
                this.closeNotification(id);
            }, duration);
        }

        return id;
    }

    closeNotification(id) {
        const notification = document.getElementById(id);
        if (notification) {
            notification.style.animation = 'notification-slide-out 0.3s ease-out';
            setTimeout(() => {
                notification.remove();
                this.notifications = this.notifications.filter(n => n !== id);
            }, 300);
        }
    }

    closeAllNotifications() {
        this.notifications.forEach(id => this.closeNotification(id));
    }

    // Enhanced Loading States
    showLoading(element, text = 'Загрузка...') {
        if (typeof element === 'string') {
            element = document.querySelector(element);
        }
        
        if (!element) return;

        element.classList.add('btn-loading');
        element.disabled = true;
        
        if (!element.dataset.originalText) {
            element.dataset.originalText = element.textContent;
        }
        
        element.innerHTML = `<span>${text}</span>`;
    }

    hideLoading(element) {
        if (typeof element === 'string') {
            element = document.querySelector(element);
        }
        
        if (!element) return;

        element.classList.remove('btn-loading');
        element.disabled = false;
        
        if (element.dataset.originalText) {
            element.textContent = element.dataset.originalText;
            delete element.dataset.originalText;
        }
    }

    // Form Validation Enhancement
    setupFormValidation() {
        document.addEventListener('input', (e) => {
            if (e.target && e.target.matches && e.target.matches('input, textarea, select')) {
                this.validateField(e.target);
            }
        });

        document.addEventListener('blur', (e) => {
            if (e.target && e.target.matches && e.target.matches('input, textarea, select')) {
                this.validateField(e.target);
            }
        });
    }

    validateField(field) {
        const container = field.closest('.form-group-enhanced, .input-enhanced');
        if (!container) return;

        // Remove existing validation states
        container.classList.remove('error', 'success');
        
        // Remove existing messages
        const existingMessages = container.querySelectorAll('.form-error, .form-success');
        existingMessages.forEach(msg => msg.remove());

        // Basic validation
        let isValid = true;
        let message = '';

        // Required field validation
        if (field.hasAttribute('required') && !field.value.trim()) {
            isValid = false;
            message = 'Это поле обязательно для заполнения';
        }

        // Email validation
        if (field.type === 'email' && field.value) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(field.value)) {
                isValid = false;
                message = 'Введите корректный email адрес';
            }
        }

        // URL validation
        if (field.type === 'url' && field.value) {
            try {
                new URL(field.value);
            } catch {
                isValid = false;
                message = 'Введите корректный URL';
            }
        }

        // Custom validation
        if (field.dataset.validate) {
            const customValidation = this.customValidations[field.dataset.validate];
            if (customValidation) {
                const result = customValidation(field.value);
                if (!result.valid) {
                    isValid = false;
                    message = result.message;
                }
            }
        }

        // Apply validation state
        if (field.value.trim()) {
            if (isValid) {
                container.classList.add('success');
                if (field.dataset.showSuccess !== 'false') {
                    this.showFieldMessage(container, 'Корректно', 'success');
                }
            } else {
                container.classList.add('error');
                this.showFieldMessage(container, message, 'error');
            }
        }
    }

    showFieldMessage(container, message, type) {
        const messageElement = document.createElement('div');
        messageElement.className = `form-${type}`;
        messageElement.innerHTML = `
            <i data-lucide="${type === 'error' ? 'alert-circle' : 'check-circle'}"></i>
            ${message}
        `;
        container.appendChild(messageElement);
        
        if (window.lucide) {
            window.lucide.createIcons();
        }
    }

    // Custom validations
    customValidations = {
        json: (value) => {
            try {
                JSON.parse(value);
                return { valid: true };
            } catch {
                return { valid: false, message: 'Некорректный JSON формат' };
            }
        },
        webhook_token: (value) => {
            if (value.length < 32) {
                return { valid: false, message: 'Токен должен содержать минимум 32 символа' };
            }
            return { valid: true };
        }
    };

    // Loading States Setup
    setupLoadingStates() {
        // Auto-handle form submissions
        document.addEventListener('submit', (e) => {
            const form = e.target;
            const submitBtn = form.querySelector('button[type="submit"], input[type="submit"]');
            
            if (submitBtn && !form.dataset.noAutoLoading) {
                this.showLoading(submitBtn, 'Сохранение...');
                
                // Reset loading state after 10 seconds (fallback)
                setTimeout(() => {
                    this.hideLoading(submitBtn);
                }, 10000);
            }
        });

        // Auto-handle AJAX requests
        const originalFetch = window.fetch;
        window.fetch = (...args) => {
            const loadingElements = document.querySelectorAll('[data-loading-for-fetch]');
            loadingElements.forEach(el => this.showLoading(el));
            
            return originalFetch(...args).finally(() => {
                loadingElements.forEach(el => this.hideLoading(el));
            });
        };
    }

    // Tooltips Setup
    setupTooltips() {
        // Auto-initialize tooltips
        document.addEventListener('mouseenter', (e) => {
            if (e.target.hasAttribute('data-tooltip') && !e.target.classList.contains('tooltip-enhanced')) {
                e.target.classList.add('tooltip-enhanced');
            }
        });
    }

    // Keyboard Shortcuts
    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + S to save
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                const saveBtn = document.querySelector('button[type="submit"], .btn-save');
                if (saveBtn && !saveBtn.disabled) {
                    saveBtn.click();
                    this.showNotification('Сохранение...', 'info', { duration: 1000 });
                }
            }

            // Escape to close modals/notifications
            if (e.key === 'Escape') {
                // Close notifications
                this.closeAllNotifications();
                
                // Close modals
                const modals = document.querySelectorAll('.modal[style*="display: flex"], .modal[style*="display: block"]');
                modals.forEach(modal => {
                    const closeBtn = modal.querySelector('.modal-close, [onclick*="close"]');
                    if (closeBtn) closeBtn.click();
                });
            }

            // Ctrl/Cmd + K for search (if search exists)
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                const searchInput = document.querySelector('input[type="search"], .search-input');
                if (searchInput) {
                    searchInput.focus();
                }
            }
        });
    }

    // Auto-save functionality
    setupAutoSave() {
        const autoSaveFields = document.querySelectorAll('[data-auto-save]');
        
        autoSaveFields.forEach(field => {
            let timeout;
            
            field.addEventListener('input', () => {
                clearTimeout(timeout);
                
                // Show saving indicator
                const indicator = this.getOrCreateAutoSaveIndicator(field);
                indicator.textContent = 'Сохранение...';
                indicator.className = 'auto-save-indicator saving';
                
                timeout = setTimeout(() => {
                    this.performAutoSave(field);
                }, 2000); // Save after 2 seconds of inactivity
            });
        });
    }

    getOrCreateAutoSaveIndicator(field) {
        let indicator = field.parentNode.querySelector('.auto-save-indicator');
        
        if (!indicator) {
            indicator = document.createElement('div');
            indicator.className = 'auto-save-indicator';
            indicator.style.cssText = `
                font-size: 0.75rem;
                color: var(--muted-foreground);
                margin-top: 0.25rem;
                transition: all 0.2s ease;
            `;
            field.parentNode.appendChild(indicator);
        }
        
        return indicator;
    }

    performAutoSave(field) {
        const indicator = this.getOrCreateAutoSaveIndicator(field);
        const saveUrl = field.dataset.autoSave;
        const fieldName = field.name || field.id;
        
        if (!saveUrl || !fieldName) return;
        
        const data = new FormData();
        data.append(fieldName, field.value);
        
        fetch(saveUrl, {
            method: 'POST',
            body: data
        })
        .then(response => {
            if (response.ok) {
                indicator.textContent = 'Сохранено';
                indicator.className = 'auto-save-indicator saved';
                indicator.style.color = 'var(--success)';
                
                setTimeout(() => {
                    indicator.textContent = '';
                }, 2000);
            } else {
                throw new Error('Save failed');
            }
        })
        .catch(() => {
            indicator.textContent = 'Ошибка сохранения';
            indicator.className = 'auto-save-indicator error';
            indicator.style.color = 'var(--destructive)';
        });
    }

    // Confirm Dialogs
    setupConfirmDialogs() {
        document.addEventListener('click', (e) => {
            if (e.target.hasAttribute('data-confirm')) {
                e.preventDefault();
                
                const message = e.target.dataset.confirm;
                const title = e.target.dataset.confirmTitle || 'Подтверждение';
                
                this.showConfirmDialog(title, message, () => {
                    // Execute original action
                    if (e.target.onclick) {
                        e.target.onclick();
                    } else if (e.target.href) {
                        window.location.href = e.target.href;
                    } else if (e.target.type === 'submit') {
                        e.target.form.submit();
                    }
                });
            }
        });
    }

    showConfirmDialog(title, message, onConfirm, onCancel = null) {
        const dialog = document.createElement('div');
        dialog.className = 'modal';
        dialog.style.display = 'flex';
        dialog.innerHTML = `
            <div class="modal-content" style="max-width: 400px;">
                <div class="modal-header">
                    <h2>${title}</h2>
                </div>
                <div class="modal-body">
                    <p>${message}</p>
                </div>
                <div class="modal-footer" style="display: flex; gap: 0.75rem; justify-content: flex-end;">
                    <button class="btn btn-secondary" onclick="this.closest('.modal').remove(); ${onCancel ? onCancel.toString() + '()' : ''}">
                        Отмена
                    </button>
                    <button class="btn btn-primary" onclick="this.closest('.modal').remove(); (${onConfirm.toString()})()">
                        Подтвердить
                    </button>
                </div>
            </div>
        `;
        
        document.body.appendChild(dialog);
        
        // Close on outside click
        dialog.addEventListener('click', (e) => {
            if (e.target === dialog) {
                dialog.remove();
                if (onCancel) onCancel();
            }
        });
    }

    // Utility Methods
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    throttle(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    // Animate elements into view
    animateOnScroll() {
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.style.animation = 'fadeInUp 0.6s ease-out';
                }
            });
        });

        document.querySelectorAll('.animate-on-scroll').forEach(el => {
            observer.observe(el);
        });
    }

    // Copy to clipboard with feedback
    copyToClipboard(text, successMessage = 'Скопировано!') {
        navigator.clipboard.writeText(text).then(() => {
            this.showNotification(successMessage, 'success', { duration: 2000 });
        }).catch(() => {
            this.showNotification('Ошибка копирования', 'error');
        });
    }
}

// Initialize UI Enhancements
const uiEnhancements = new UIEnhancements();

// Global utility functions
window.showNotification = (message, type, options) => uiEnhancements.showNotification(message, type, options);
window.showLoading = (element, text) => uiEnhancements.showLoading(element, text);
window.hideLoading = (element) => uiEnhancements.hideLoading(element);
window.copyToClipboard = (text, message) => uiEnhancements.copyToClipboard(text, message);

// CSS for animations
const style = document.createElement('style');
style.textContent = `
    @keyframes fadeInUp {
        from {
            opacity: 0;
            transform: translateY(20px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
    
    .auto-save-indicator.saving {
        color: var(--primary) !important;
    }
    
    .auto-save-indicator.saved {
        color: var(--success) !important;
    }
    
    .auto-save-indicator.error {
        color: var(--destructive) !important;
    }
`;
document.head.appendChild(style);