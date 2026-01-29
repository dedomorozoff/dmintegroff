// История запросов - JavaScript функциональность
class RequestHistory {
    constructor() {
        this.currentPage = 1;
        this.perPage = 50;
        this.currentFilters = {};
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadHistory();
        this.setDefaultDates();
        
        // Initialize Lucide icons
        if (window.lucide) {
            window.lucide.createIcons();
        }
    }

    bindEvents() {
        // Обработчики для фильтров
        document.getElementById('filterSearch').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.applyFilters();
            }
        });

        // Автоматическое применение фильтров при изменении селектов
        const selects = ['filterIntegration', 'filterStatus', 'filterMethod', 'filterLogType'];
        selects.forEach(id => {
            document.getElementById(id).addEventListener('change', () => {
                this.applyFilters();
            });
        });

        // Автоматическое применение фильтров при изменении дат
        document.getElementById('filterDateFrom').addEventListener('change', () => {
            this.applyFilters();
        });
        document.getElementById('filterDateTo').addEventListener('change', () => {
            this.applyFilters();
        });
    }

    setDefaultDates() {
        // Устанавливаем дату "от" на неделю назад
        const dateFrom = new Date();
        dateFrom.setDate(dateFrom.getDate() - 7);
        document.getElementById('filterDateFrom').value = dateFrom.toISOString().split('T')[0];

        // Устанавливаем дату "до" на сегодня
        const dateTo = new Date();
        document.getElementById('filterDateTo').value = dateTo.toISOString().split('T')[0];
    }

    getFilters() {
        return {
            integration_id: document.getElementById('filterIntegration').value,
            status: document.getElementById('filterStatus').value,
            method: document.getElementById('filterMethod').value,
            log_type: document.getElementById('filterLogType').value,
            date_from: document.getElementById('filterDateFrom').value,
            date_to: document.getElementById('filterDateTo').value,
            search: document.getElementById('filterSearch').value.trim(),
            output_name: document.getElementById('filterOutputName').value.trim()
        };
    }

    async loadHistory(page = 1) {
        this.currentPage = page;
        this.showLoading(true);

        try {
            const filters = this.getFilters();
            this.currentFilters = filters;

            const params = new URLSearchParams({
                page: page.toString(),
                per_page: this.perPage.toString(),
                ...filters
            });

            const response = await fetch(`/api/logs/requests?${params}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            const data = await response.json();
            this.renderHistory(data);
            this.renderPagination(data);
            this.updatePaginationInfo(data);

        } catch (error) {
            console.error('Ошибка загрузки истории:', error);
            this.showError('Ошибка загрузки данных. Попробуйте обновить страницу.');
        } finally {
            this.showLoading(false);
        }
    }

    renderHistory(data) {
        const tbody = document.getElementById('requestsTableBody');
        const emptyState = document.getElementById('emptyState');
        const resultsContainer = document.getElementById('resultsContainer');

        if (!data.requests || data.requests.length === 0) {
            tbody.innerHTML = '';
            resultsContainer.style.display = 'none';
            emptyState.style.display = 'block';
            return;
        }

        resultsContainer.style.display = 'block';
        emptyState.style.display = 'none';

        tbody.innerHTML = data.requests.map(request => {
            const statusClass = request.status_code >= 400 ? 'status-error' : 'status-success';
            const methodClass = this.getMethodClass(request.method);
            const createdAt = new Date(request.created_at).toLocaleString('ru-RU');
            
            const requestSize = this.formatBytes(request.request_size);
            const responseSize = this.formatBytes(request.response_size);
            
            return `
                <tr>
                    <td>
                        <div style="font-size: 0.85rem;">${createdAt}</div>
                    </td>
                    <td>
                        <div style="font-weight: 500;">${this.escapeHtml(request.integration_name || 'N/A')}</div>
                        <div style="font-size: 0.8rem; color: var(--text-secondary);">ID: ${request.integration_id}</div>
                    </td>
                    <td>
                        <span class="request-method-badge ${methodClass}">
                            ${request.method}
                        </span>
                    </td>
                    <td>
                        <div class="request-url" title="${this.escapeHtml(request.url)}">
                            ${this.escapeHtml(request.url)}
                        </div>
                    </td>
                    <td>
                        <span class="${statusClass}" style="font-weight: 600;">
                            ${request.status_code}
                        </span>
                        ${request.error_message ? `<div style="font-size: 0.8rem; color: var(--destructive); margin-top: 0.25rem;">${this.escapeHtml(request.error_message.substring(0, 50))}${request.error_message.length > 50 ? '...' : ''}</div>` : ''}
                    </td>
                    <td>
                        <span class="status-badge secondary">${request.log_type}</span>
                    </td>
                    <td>
                        ${request.output_name ? `<div style="font-size: 0.8rem; color: var(--text-secondary);">${this.escapeHtml(request.output_name)}</div>` : '-'}
                    </td>
                    <td>
                        <div class="size-info">
                            ${requestSize}${responseSize ? ` / ${responseSize}` : ''}
                        </div>
                    </td>
                    <td>
                        <button type="button" class="btn btn-sm btn-secondary" onclick="requestHistory.showDetails(${request.id})">
                            <i data-lucide="eye"></i>
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
        
        // Re-initialize Lucide icons for new content
        if (window.lucide) {
            window.lucide.createIcons();
        }
    }

    renderPagination(data) {
        const paginationNav = document.getElementById('paginationNav');
        
        if (data.total_pages <= 1) {
            paginationNav.innerHTML = '';
            return;
        }

        let paginationHTML = '';

        // Previous button
        paginationHTML += `
            <button ${!data.has_prev ? 'disabled' : ''} onclick="requestHistory.loadHistory(${data.page - 1})">
                <i data-lucide="chevron-left"></i>
            </button>
        `;

        // Page numbers
        const startPage = Math.max(1, data.page - 2);
        const endPage = Math.min(data.total_pages, data.page + 2);

        if (startPage > 1) {
            paginationHTML += `<button onclick="requestHistory.loadHistory(1)">1</button>`;
            if (startPage > 2) {
                paginationHTML += '<span style="padding: 0.5rem;">...</span>';
            }
        }

        for (let i = startPage; i <= endPage; i++) {
            paginationHTML += `
                <button ${i === data.page ? 'class="active"' : ''} onclick="requestHistory.loadHistory(${i})">
                    ${i}
                </button>
            `;
        }

        if (endPage < data.total_pages) {
            if (endPage < data.total_pages - 1) {
                paginationHTML += '<span style="padding: 0.5rem;">...</span>';
            }
            paginationHTML += `<button onclick="requestHistory.loadHistory(${data.total_pages})">${data.total_pages}</button>`;
        }

        // Next button
        paginationHTML += `
            <button ${!data.has_next ? 'disabled' : ''} onclick="requestHistory.loadHistory(${data.page + 1})">
                <i data-lucide="chevron-right"></i>
            </button>
        `;

        paginationNav.innerHTML = paginationHTML;
        
        // Re-initialize Lucide icons
        if (window.lucide) {
            window.lucide.createIcons();
        }
    }

    updatePaginationInfo(data) {
        const paginationInfo = document.getElementById('paginationInfo');
        const start = (data.page - 1) * data.per_page + 1;
        const end = Math.min(data.page * data.per_page, data.total);
        
        paginationInfo.innerHTML = `
            Показано ${start}-${end} из ${data.total} записей
            ${data.total_pages > 1 ? `(страница ${data.page} из ${data.total_pages})` : ''}
        `;
    }

    async showDetails(requestId) {
        try {
            const response = await fetch(`/api/logs/requests/${requestId}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            const request = await response.json();
            this.renderRequestDetails(request);
            
            const modal = document.getElementById('requestDetailsModal');
            modal.style.display = 'flex';

        } catch (error) {
            console.error('Ошибка загрузки деталей:', error);
            this.showError('Ошибка загрузки деталей запроса');
        }
    }

    renderRequestDetails(request) {
        const modalBody = document.getElementById('requestDetailsBody');
        const createdAt = new Date(request.created_at).toLocaleString('ru-RU');
        const statusClass = request.status_code >= 400 ? 'status-error' : 'status-success';
        const methodClass = this.getMethodClass(request.method);

        let requestHeaders = '';
        if (request.request_headers && Object.keys(request.request_headers).length > 0) {
            requestHeaders = Object.entries(request.request_headers)
                .map(([key, values]) => `<tr><td><strong>${this.escapeHtml(key)}</strong></td><td>${this.escapeHtml(Array.isArray(values) ? values.join(', ') : values)}</td></tr>`)
                .join('');
        }

        modalBody.innerHTML = `
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem; margin-bottom: 2rem;">
                <div>
                    <h3 style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem;">
                        <i data-lucide="info"></i>
                        Основная информация
                    </h3>
                    <div style="background: var(--secondary); padding: 1rem; border-radius: 8px;">
                        <div style="margin-bottom: 0.75rem;"><strong>ID:</strong> ${request.id}</div>
                        <div style="margin-bottom: 0.75rem;"><strong>Дата/Время:</strong> ${createdAt}</div>
                        <div style="margin-bottom: 0.75rem;"><strong>Интеграция:</strong> ${this.escapeHtml(request.integration_name || 'N/A')} (ID: ${request.integration_id})</div>
                        <div style="margin-bottom: 0.75rem;"><strong>Метод:</strong> <span class="request-method-badge ${methodClass}">${request.method}</span></div>
                        <div style="margin-bottom: 0.75rem;"><strong>URL:</strong> <div style="word-break: break-all; font-family: monospace; font-size: 0.9rem; margin-top: 0.25rem;">${this.escapeHtml(request.url)}</div></div>
                        <div style="margin-bottom: 0.75rem;"><strong>Статус:</strong> <span class="${statusClass}" style="font-weight: 600;">${request.status_code}</span></div>
                        <div style="margin-bottom: 0.75rem;"><strong>Тип лога:</strong> <span class="status-badge secondary">${request.log_type}</span></div>
                        ${request.output_name ? `<div><strong>Название выхода:</strong> ${this.escapeHtml(request.output_name)}</div>` : ''}
                    </div>
                </div>
                <div>
                    <h3 style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem;">
                        <i data-lucide="bar-chart-3"></i>
                        Статистика
                    </h3>
                    <div style="background: var(--secondary); padding: 1rem; border-radius: 8px;">
                        <div style="margin-bottom: 0.75rem;"><strong>Размер запроса:</strong> ${this.formatBytes(request.request_body ? request.request_body.length : 0)}</div>
                        <div style="margin-bottom: 0.75rem;"><strong>Размер ответа:</strong> ${this.formatBytes(request.response_body ? request.response_body.length : 0)}</div>
                        <div style="margin-bottom: 0.75rem;"><strong>Есть заголовки:</strong> ${request.request_headers ? '<i data-lucide="check" style="color: var(--success);"></i>' : '<i data-lucide="x" style="color: var(--text-secondary);"></i>'}</div>
                        <div style="margin-bottom: 0.75rem;"><strong>Есть тело запроса:</strong> ${request.request_body ? '<i data-lucide="check" style="color: var(--success);"></i>' : '<i data-lucide="x" style="color: var(--text-secondary);"></i>'}</div>
                        <div><strong>Есть тело ответа:</strong> ${request.response_body ? '<i data-lucide="check" style="color: var(--success);"></i>' : '<i data-lucide="x" style="color: var(--text-secondary);"></i>'}</div>
                    </div>
                </div>
            </div>

            ${request.error_message ? `
                <div style="margin-bottom: 2rem;">
                    <h3 style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem; color: var(--destructive);">
                        <i data-lucide="alert-triangle"></i>
                        Ошибка
                    </h3>
                    <div style="background: color-mix(in srgb, var(--destructive) 10%, transparent); border: 1px solid var(--destructive); border-radius: 8px; padding: 1rem;">
                        <pre style="margin: 0; white-space: pre-wrap; font-family: monospace; font-size: 0.9rem;">${this.escapeHtml(request.error_message)}</pre>
                    </div>
                </div>
            ` : ''}

            ${requestHeaders ? `
                <div style="margin-bottom: 2rem;">
                    <h3 style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem;">
                        <i data-lucide="list"></i>
                        Заголовки запроса
                    </h3>
                    <div style="overflow-x: auto;">
                        <table class="headers-table" style="width: 100%; border-collapse: collapse;">
                            <thead>
                                <tr style="background: var(--secondary);">
                                    <th style="padding: 0.75rem; text-align: left; border-bottom: 1px solid var(--border);">Заголовок</th>
                                    <th style="padding: 0.75rem; text-align: left; border-bottom: 1px solid var(--border);">Значение</th>
                                </tr>
                            </thead>
                            <tbody>${requestHeaders}</tbody>
                        </table>
                    </div>
                </div>
            ` : ''}

            ${request.request_body ? `
                <div style="margin-bottom: 2rem;">
                    <h3 style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem;">
                        <i data-lucide="file-text"></i>
                        Тело запроса
                    </h3>
                    <div class="json-viewer">${this.formatJSON(request.request_body)}</div>
                </div>
            ` : ''}

            ${request.response_body ? `
                <div style="margin-bottom: 2rem;">
                    <h3 style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem;">
                        <i data-lucide="corner-down-left"></i>
                        Тело ответа
                    </h3>
                    <div class="json-viewer">${this.formatJSON(request.response_body)}</div>
                </div>
            ` : ''}
        `;
        
        // Re-initialize Lucide icons
        if (window.lucide) {
            window.lucide.createIcons();
        }
    }

    applyFilters() {
        this.loadHistory(1);
    }

    clearFilters() {
        document.getElementById('filterIntegration').value = 'all';
        document.getElementById('filterStatus').value = 'all';
        document.getElementById('filterMethod').value = 'all';
        document.getElementById('filterLogType').value = 'all';
        document.getElementById('filterSearch').value = '';
        document.getElementById('filterOutputName').value = '';
        
        // Сбрасываем даты на значения по умолчанию
        this.setDefaultDates();
        
        this.loadHistory(1);
    }

    refreshHistory() {
        this.loadHistory(this.currentPage);
    }

    async exportHistory() {
        try {
            const filters = this.getFilters();
            const params = new URLSearchParams(filters);
            
            const response = await fetch(`/api/logs/requests/export?${params}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            // Получаем имя файла из заголовка Content-Disposition
            const contentDisposition = response.headers.get('Content-Disposition');
            let filename = 'request_history.csv';
            if (contentDisposition) {
                const filenameMatch = contentDisposition.match(/filename="(.+)"/);
                if (filenameMatch) {
                    filename = filenameMatch[1];
                }
            }

            // Скачиваем файл
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);

            this.showSuccess('Экспорт завершен успешно');

        } catch (error) {
            console.error('Ошибка экспорта:', error);
            this.showError('Ошибка экспорта данных');
        }
    }

    // Utility methods
    showLoading(show) {
        const loading = document.getElementById('loadingIndicator');
        const results = document.getElementById('resultsContainer');
        const empty = document.getElementById('emptyState');
        
        if (show) {
            loading.style.display = 'block';
            results.style.display = 'none';
            empty.style.display = 'none';
        } else {
            loading.style.display = 'none';
        }
    }

    showError(message) {
        // Можно использовать toast или alert
        alert('Ошибка: ' + message);
    }

    showSuccess(message) {
        // Можно использовать toast
        console.log('Успех: ' + message);
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    }

    formatJSON(jsonString) {
        if (!jsonString) return '';
        
        try {
            const parsed = JSON.parse(jsonString);
            return this.escapeHtml(JSON.stringify(parsed, null, 2));
        } catch (e) {
            return this.escapeHtml(jsonString);
        }
    }

    getMethodClass(method) {
        const classes = {
            'GET': 'method-get',
            'POST': 'method-post',
            'PUT': 'method-put',
            'DELETE': 'method-delete',
            'PATCH': 'method-patch'
        };
        return classes[method] || 'method-default';
    }
    
    getMethodColor(method) {
        const colors = {
            'GET': '#10b981',
            'POST': '#3b82f6',
            'PUT': '#f59e0b',
            'DELETE': '#ef4444',
            'PATCH': '#8b5cf6'
        };
        return colors[method] || '#6b7280';
    }
}

// Глобальные функции для использования в HTML
let requestHistory;

document.addEventListener('DOMContentLoaded', function() {
    requestHistory = new RequestHistory();
});

function applyFilters() {
    requestHistory.applyFilters();
}

function clearFilters() {
    requestHistory.clearFilters();
}

function refreshHistory() {
    requestHistory.refreshHistory();
}

function exportHistory() {
    requestHistory.exportHistory();
}

function closeRequestDetails() {
    const modal = document.getElementById('requestDetailsModal');
    modal.style.display = 'none';
}

// Закрытие модального окна при клике вне его
window.onclick = function(event) {
    const modal = document.getElementById('requestDetailsModal');
    if (event.target === modal) {
        closeRequestDetails();
    }
}