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

// Enhanced Modal for AI Test Results
function showAITestResultModal(title, content, options = {}) {
        const {
            size = 'large', // small, medium, large
            type = 'info', // success, error, warning, info
            scrollable = true,
            fullHeight = false
        } = options;

        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.display = 'flex';
        
        const modalId = 'ai-test-modal-' + Date.now();
        modal.id = modalId;

        modal.innerHTML = `
            <div class="modal-content ${size} ${fullHeight ? 'full-height' : ''}">
                <div class="modal-header">
                    <h2>${title}</h2>
                    <button class="modal-close" onclick="document.getElementById('${modalId}').remove()">
                        <i data-lucide="x"></i>
                    </button>
                </div>
                <div class="modal-body ${scrollable ? 'scrollable' : ''}">
                    ${formatAITestContent(content, type)}
                </div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" onclick="document.getElementById('${modalId}').remove()">
                        Закрыть
                    </button>
                    <button class="btn btn-primary" onclick="uiEnhancements.copyAITestResult('${modalId}')">
                        <i data-lucide="copy"></i>
                        Копировать результат
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        // Initialize Lucide icons
        if (window.lucide) {
            window.lucide.createIcons();
        }

        // Close on outside click
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });

        // Close on Escape key
        const escapeHandler = (e) => {
            if (e.key === 'Escape') {
                modal.remove();
                document.removeEventListener('keydown', escapeHandler);
            }
        };
        document.addEventListener('keydown', escapeHandler);

        return modalId;
    }

function formatAITestContent(content, type) {
        if (typeof content === 'string') {
            // If it's a simple string, wrap it in a result container
            return `
                <div class="ai-test-result ${type}">
                    <div class="ai-test-result-header">
                        <i data-lucide="${getIconForType(type)}"></i>
                        Результат тестирования AI
                    </div>
                    <div class="ai-test-result-content">
                        <pre><code>${escapeHtml(content)}</code></pre>
                    </div>
                </div>
            `;
        }

        if (typeof content === 'object') {
            // If it's an object, format it as structured data
            let html = `
                <div class="ai-test-result ${type}">
                    <div class="ai-test-result-header">
                        <i data-lucide="${getIconForType(type)}"></i>
                        Результат тестирования AI
                    </div>
                    <div class="ai-test-result-content">
            `;

            // Add each section
            Object.keys(content).forEach(key => {
                const value = content[key];
                html += `
                    <div class="response-section">
                        <h4>
                            <i data-lucide="${getIconForSection(key)}"></i>
                            ${formatSectionTitle(key)}
                        </h4>
                        <div class="section-content">
                            ${formatSectionContent(value)}
                        </div>
                    </div>
                `;
            });

            html += `
                    </div>
                </div>
            `;

            return html;
        }

        return content;
    }

function getIconForType(type) {
        const icons = {
            success: 'check-circle',
            error: 'alert-circle',
            warning: 'alert-triangle',
            info: 'info'
        };
        return icons[type] || 'info';
    }

function getIconForSection(key) {
        const icons = {
            request: 'send',
            response: 'arrow-left',
            error: 'alert-circle',
            data: 'database',
            mapping: 'shuffle',
            result: 'check-circle',
            logs: 'file-text',
            config: 'settings'
        };
        return icons[key.toLowerCase()] || 'chevron-right';
    }

function formatSectionTitle(key) {
        const titles = {
            request: 'Запрос',
            response: 'Ответ',
            error: 'Ошибка',
            data: 'Данные',
            mapping: 'Маппинг',
            result: 'Результат',
            logs: 'Логи',
            config: 'Конфигурация'
        };
        return titles[key.toLowerCase()] || key.charAt(0).toUpperCase() + key.slice(1);
    }

function formatSectionContent(value) {
        if (typeof value === 'object') {
            return `<div class="json-content">${JSON.stringify(value, null, 2)}</div>`;
        }
        
        if (typeof value === 'string' && isJSON(value)) {
            try {
                const parsed = JSON.parse(value);
                return `<div class="json-content">${JSON.stringify(parsed, null, 2)}</div>`;
            } catch (e) {
                return `<pre><code>${escapeHtml(value)}</code></pre>`;
            }
        }
        
        return `<pre><code>${escapeHtml(String(value))}</code></pre>`;
    }

function isJSON(str) {
        try {
            JSON.parse(str);
            return true;
        } catch (e) {
            return false;
        }
    }

function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

function copyAITestResult(modalId) {
        const modal = document.getElementById(modalId);
        if (!modal) return;

        const content = modal.querySelector('.ai-test-result-content');
        if (!content) return;

        // Extract text content
        let textContent = '';
        
        const sections = content.querySelectorAll('.response-section');
        if (sections.length > 0) {
            sections.forEach(section => {
                const title = section.querySelector('h4').textContent.trim();
                const sectionContent = section.querySelector('.section-content');
                const text = sectionContent.textContent.trim();
                
                textContent += `${title}:\n${text}\n\n`;
            });
        } else {
            textContent = content.textContent.trim();
        }

        uiEnhancements.copyToClipboard(textContent, 'Результат скопирован в буфер обмена');
    }

// Quick method to show AI test results
function showAITestResult(data, type = 'info') {
        let title = 'Результат тестирования AI';
        
        switch (type) {
            case 'success':
                title = '✅ Тестирование успешно завершено';
                break;
            case 'error':
                title = '❌ Ошибка при тестировании';
                break;
            case 'warning':
                title = '⚠️ Тестирование завершено с предупреждениями';
                break;
        }

        return showAITestResultModal(title, data, {
            type: type,
            size: 'large',
            scrollable: true,
            fullHeight: true
        });
    }

// Global functions for AI test results
window.showAITestResult = (data, type) => uiEnhancements.showAITestResult(data, type);
window.showAITestResultModal = (title, content, options) => uiEnhancements.showAITestResultModal(title, content, options);

// Example usage:
/*
// Simple text result
showAITestResult("Тестирование прошло успешно!", "success");

// Complex object result
showAITestResult({
    request: {
        url: "https://api.example.com/webhook",
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: '{"test": true}'
    },
    response: {
        status: 200,
        body: '{"success": true, "id": 123}'
    },
    mapping: {
        "id": "response.id",
        "status": "success"
    }
}, "success");

// Error result
showAITestResult({
    error: "Connection timeout",
    logs: "Failed to connect to https://api.example.com after 30 seconds"
}, "error");
*/

// Sidebar Submenu Management
class SidebarSubmenuManager {
    constructor() {
        this.init();
    }

    init() {
        this.setupSubmenuToggle();
        this.setupActiveStates();
    }

    setupSubmenuToggle() {
        document.addEventListener('DOMContentLoaded', () => {
            const submenuItems = document.querySelectorAll('.sidebar-submenu');
            
            submenuItems.forEach(item => {
                const mainLink = item.querySelector('> a');
                const submenu = item.querySelector('.submenu');
                
                if (mainLink && submenu) {
                    // Показываем подменю при наведении
                    item.addEventListener('mouseenter', () => {
                        this.showSubmenu(item);
                    });
                    
                    // Скрываем подменю при уходе мыши (с задержкой)
                    item.addEventListener('mouseleave', () => {
                        setTimeout(() => {
                            if (!item.matches(':hover')) {
                                this.hideSubmenu(item);
                            }
                        }, 300);
                    });
                    
                    // Клик по основной ссылке переключает подменю
                    mainLink.addEventListener('click', (e) => {
                        if (window.innerWidth <= 768) { // Мобильная версия
                            e.preventDefault();
                            this.toggleSubmenu(item);
                        }
                    });
                }
            });
        });
    }

    setupActiveStates() {
        document.addEventListener('DOMContentLoaded', () => {
            // Автоматически показываем подменю если один из его элементов активен
            const activeSubmenuItems = document.querySelectorAll('.sidebar-submenu .submenu a.active');
            
            activeSubmenuItems.forEach(activeItem => {
                const submenuContainer = activeItem.closest('.sidebar-submenu');
                if (submenuContainer) {
                    submenuContainer.classList.add('active');
                    this.showSubmenu(submenuContainer, true);
                }
            });
        });
    }

    showSubmenu(submenuItem, permanent = false) {
        const submenu = submenuItem.querySelector('.submenu');
        if (submenu) {
            submenu.style.maxHeight = submenu.scrollHeight + 'px';
            submenu.style.opacity = '1';
            
            if (permanent) {
                submenuItem.classList.add('active');
            }
        }
    }

    hideSubmenu(submenuItem) {
        // Не скрываем если элемент активен
        if (submenuItem.classList.contains('active')) {
            return;
        }
        
        const submenu = submenuItem.querySelector('.submenu');
        if (submenu) {
            submenu.style.maxHeight = '0';
            submenu.style.opacity = '0';
        }
    }

    toggleSubmenu(submenuItem) {
        const submenu = submenuItem.querySelector('.submenu');
        if (!submenu) return;
        
        const isOpen = submenu.style.maxHeight && submenu.style.maxHeight !== '0px';
        
        if (isOpen) {
            this.hideSubmenu(submenuItem);
            submenuItem.classList.remove('active');
        } else {
            this.showSubmenu(submenuItem, true);
            submenuItem.classList.add('active');
        }
    }
}

// Initialize Sidebar Submenu Manager
const sidebarSubmenuManager = new SidebarSubmenuManager();

// Add CSS for submenu animations
const submenuStyle = document.createElement('style');
submenuStyle.textContent = `
    .ai-test-result {
        border: 1px solid var(--border);
        border-radius: var(--radius);
        overflow: hidden;
        margin-bottom: 1rem;
    }

    .ai-test-result.success {
        border-color: var(--success);
        background: color-mix(in srgb, var(--success) 5%, var(--card));
    }

    .ai-test-result.error {
        border-color: var(--destructive);
        background: color-mix(in srgb, var(--destructive) 5%, var(--card));
    }

    .ai-test-result.warning {
        border-color: var(--warning);
        background: color-mix(in srgb, var(--warning) 5%, var(--card));
    }

    .ai-test-result-header {
        background: var(--muted);
        padding: 1rem;
        font-weight: 600;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        border-bottom: 1px solid var(--border);
    }

    .ai-test-result-content {
        padding: 1rem;
    }

    .response-section {
        margin-bottom: 1.5rem;
    }

    .response-section:last-child {
        margin-bottom: 0;
    }

    .response-section h4 {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        margin: 0 0 0.75rem 0;
        color: var(--foreground);
        font-size: 1rem;
        font-weight: 600;
    }

    .section-content {
        background: var(--background);
        border: 1px solid var(--border);
        border-radius: var(--radius);
        padding: 1rem;
    }

    .json-content {
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 0.875rem;
        line-height: 1.5;
        white-space: pre-wrap;
        word-break: break-all;
        background: var(--muted);
        padding: 1rem;
        border-radius: var(--radius);
        border: 1px solid var(--border);
        overflow-x: auto;
    }

    .modal-content.large {
        max-width: 800px;
        width: 90vw;
    }

    .modal-content.full-height {
        height: 90vh;
        max-height: 90vh;
    }

    .modal-body.scrollable {
        overflow-y: auto;
        max-height: calc(90vh - 120px);
    }

    /* Submenu specific styles */
    .sidebar-submenu .submenu {
        transition: max-height 0.3s cubic-bezier(0.4, 0, 0.2, 1), 
                    opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }

    .sidebar-submenu.active > a {
        background: color-mix(in srgb, var(--primary) 10%, transparent);
        color: var(--primary);
    }

    /* Mobile submenu styles */
    @media (max-width: 768px) {
        .sidebar-submenu .submenu {
            background: color-mix(in srgb, var(--background) 95%, var(--muted));
            border-radius: var(--radius);
            margin: 0.25rem 0;
            border: 1px solid var(--border);
        }
    }
`;
document.head.appendChild(submenuStyle);