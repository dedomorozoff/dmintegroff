// AI Chat Widget JavaScript
class AIChat {
    constructor() {
        this.isOpen = false;
        this.isTyping = false;
        this.messages = [];
        this.init();
    }

    init() {
        // Проверяем статус AI при загрузке
        this.checkAIStatus();
        
        // Добавляем обработчик Enter для отправки сообщений
        const input = document.getElementById('aiChatInput');
        if (input) {
            input.addEventListener('keypress', (e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    this.sendMessage();
                }
            });
        }

        // Инициализируем Lucide иконки
        if (window.lucide) {
            window.lucide.createIcons();
        }
        
        // Принудительно обновляем иконки через небольшую задержку
        setTimeout(() => {
            if (window.lucide) {
                window.lucide.createIcons();
            }
        }, 100);
    }

    async checkAIStatus() {
        try {
            const response = await fetch('/api/ai/status');
            const status = await response.json();
            
            const badge = document.getElementById('aiChatBadge');
            const statusDiv = document.getElementById('aiChatStatus');
            
            if (status.available || status.demo_mode) {
                // AI доступен или в демо режиме
                if (badge) {
                    badge.textContent = status.demo_mode ? 'DEMO' : 'AI';
                    badge.style.background = status.demo_mode ? 'var(--warning)' : 'var(--success)';
                }
                
                if (statusDiv) {
                    statusDiv.style.display = 'none';
                }
                
                // Обновляем приветственное сообщение
                this.updateWelcomeMessage(status);
            } else {
                // AI недоступен
                if (badge) {
                    badge.textContent = 'OFF';
                    badge.style.background = 'var(--destructive)';
                }
                
                if (statusDiv) {
                    statusDiv.innerHTML = `
                        <div class="ai-status-content">
                            <i data-lucide="alert-circle" style="color: var(--destructive);"></i>
                            <span>${status.error || 'AI недоступен'}</span>
                        </div>
                    `;
                    statusDiv.style.display = 'block';
                }
            }
            
            if (window.lucide) {
                window.lucide.createIcons();
            }
        } catch (error) {
            console.error('Ошибка проверки статуса AI:', error);
            
            const badge = document.getElementById('aiChatBadge');
            if (badge) {
                badge.textContent = 'ERR';
                badge.style.background = 'var(--destructive)';
            }
        }
    }

    updateWelcomeMessage(status) {
        const messagesContainer = document.getElementById('aiChatMessages');
        if (!messagesContainer) return;

        let welcomeText = 'Привет! Я помогу создать интеграцию. Просто опишите что нужно сделать.';
        
        if (status.demo_mode) {
            welcomeText = '🎮 Демо режим активен! Я покажу примеры того, как AI может помочь с интеграциями. Попробуйте спросить про Slack, Telegram или другие популярные сервисы.';
        }

        // Обновляем первое сообщение
        const firstMessage = messagesContainer.querySelector('.ai-message .ai-message-text');
        if (firstMessage) {
            firstMessage.textContent = welcomeText;
        }
    }

    toggle() {
        const window = document.getElementById('aiChatWindow');
        const toggle = document.getElementById('aiChatToggle');
        
        if (!window || !toggle) return;

        if (this.isOpen) {
            this.close();
        } else {
            this.open();
        }
    }

    open() {
        const window = document.getElementById('aiChatWindow');
        const input = document.getElementById('aiChatInput');
        
        if (!window) return;

        window.classList.add('open');
        this.isOpen = true;
        
        // Фокус на поле ввода
        setTimeout(() => {
            if (input) input.focus();
        }, 300);

        // Прокручиваем к последнему сообщению
        this.scrollToBottom();
    }

    close() {
        const window = document.getElementById('aiChatWindow');
        
        if (!window) return;

        window.classList.remove('open');
        this.isOpen = false;
    }

    async sendMessage(text = null) {
        const input = document.getElementById('aiChatInput');
        const sendButton = document.getElementById('aiSendButton');
        
        if (!input || !sendButton) return;

        const message = text || input.value.trim();
        if (!message) return;

        // Очищаем поле ввода
        if (!text) input.value = '';
        
        // Блокируем кнопку отправки
        sendButton.disabled = true;

        // Добавляем сообщение пользователя
        this.addUserMessage(message);

        // Показываем индикатор печати
        this.showTyping();

        try {
            // Отправляем запрос к AI
            const response = await fetch('/api/ai/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    message: message,
                    context: this.getPageContext(),
                    history: this.getRecentHistory()
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            const data = await response.json();
            
            // Скрываем индикатор печати
            this.hideTyping();
            
            // Добавляем ответ AI
            this.addAIMessage(data.response);
            
            // Показываем предложения если есть
            if (data.suggestions && data.suggestions.length > 0) {
                this.showSuggestions(data.suggestions);
            }

        } catch (error) {
            console.error('Ошибка отправки сообщения:', error);
            
            this.hideTyping();
            
            let errorMessage = 'Извините, произошла ошибка. Попробуйте еще раз.';
            
            if (error.message.includes('503')) {
                errorMessage = 'AI временно недоступен. Проверьте настройки в разделе Настройки.';
            } else if (error.message.includes('401')) {
                errorMessage = 'Необходимо войти в систему для использования AI.';
            }
            
            this.addAIMessage(errorMessage);
        } finally {
            // Разблокируем кнопку отправки
            sendButton.disabled = false;
        }
    }

    sendQuickMessage(text) {
        this.sendMessage(text);
    }

    addUserMessage(text) {
        const messagesContainer = document.getElementById('aiChatMessages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = 'user-message';
        messageDiv.innerHTML = `
            <div class="user-message-content">
                <div class="user-message-text">${this.escapeHtml(text)}</div>
                <div class="user-message-time">${this.getCurrentTime()}</div>
            </div>
            <div class="user-avatar">👤</div>
        `;

        messagesContainer.appendChild(messageDiv);
        this.scrollToBottom();

        // Сохраняем в историю
        this.messages.push({
            role: 'user',
            content: text,
            timestamp: Date.now()
        });
    }

    addAIMessage(text) {
        const messagesContainer = document.getElementById('aiChatMessages');
        if (!messagesContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = 'ai-message';
        messageDiv.innerHTML = `
            <div class="ai-avatar">🤖</div>
            <div class="ai-message-content">
                <div class="ai-message-text">${this.formatAIResponse(text)}</div>
                <div class="ai-message-time">${this.getCurrentTime()}</div>
            </div>
        `;

        messagesContainer.appendChild(messageDiv);
        this.scrollToBottom();

        // Сохраняем в историю
        this.messages.push({
            role: 'assistant',
            content: text,
            timestamp: Date.now()
        });
    }

    showTyping() {
        const messagesContainer = document.getElementById('aiChatMessages');
        if (!messagesContainer || this.isTyping) return;

        this.isTyping = true;

        const typingDiv = document.createElement('div');
        typingDiv.className = 'ai-message typing-message';
        typingDiv.id = 'typingIndicator';
        typingDiv.innerHTML = `
            <div class="ai-avatar">🤖</div>
            <div class="ai-message-content">
                <div class="ai-typing">
                    <div class="ai-typing-dot"></div>
                    <div class="ai-typing-dot"></div>
                    <div class="ai-typing-dot"></div>
                </div>
            </div>
        `;

        messagesContainer.appendChild(typingDiv);
        this.scrollToBottom();
    }

    hideTyping() {
        const typingIndicator = document.getElementById('typingIndicator');
        if (typingIndicator) {
            typingIndicator.remove();
        }
        this.isTyping = false;
    }

    showSuggestions(suggestions) {
        const messagesContainer = document.getElementById('aiChatMessages');
        if (!messagesContainer) return;

        const suggestionsDiv = document.createElement('div');
        suggestionsDiv.className = 'ai-message suggestions-message';
        
        let suggestionsHtml = suggestions.slice(0, 3).map(suggestion => `
            <button class="ai-suggestion-btn" onclick="aiChat.sendQuickMessage('${this.escapeHtml(suggestion.title)}')">
                ${this.escapeHtml(suggestion.title)}
            </button>
        `).join('');

        suggestionsDiv.innerHTML = `
            <div class="ai-avatar">💡</div>
            <div class="ai-message-content">
                <div class="ai-suggestions">
                    <div class="ai-suggestions-title">Предложения:</div>
                    <div class="ai-suggestions-buttons">
                        ${suggestionsHtml}
                    </div>
                </div>
            </div>
        `;

        messagesContainer.appendChild(suggestionsDiv);
        this.scrollToBottom();
    }

    formatAIResponse(text) {
        // Простое форматирование ответа AI
        return text
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>') // **bold**
            .replace(/\*(.*?)\*/g, '<em>$1</em>') // *italic*
            .replace(/`(.*?)`/g, '<code>$1</code>') // `code`
            .replace(/\n/g, '<br>'); // переносы строк
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    getCurrentTime() {
        const now = new Date();
        return now.toLocaleTimeString('ru-RU', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
    }

    getPageContext() {
        // Получаем контекст текущей страницы
        const context = {
            page: window.location.pathname,
            title: document.title
        };

        // Добавляем специфичную информацию в зависимости от страницы
        if (context.page === '/dashboard') {
            context.type = 'dashboard';
            context.description = 'Пользователь на главной странице';
        }

        return context;
    }

    getRecentHistory() {
        // Возвращаем последние 5 сообщений для контекста
        return this.messages.slice(-5).map(msg => ({
            role: msg.role,
            content: msg.content
        }));
    }

    scrollToBottom() {
        const messagesContainer = document.getElementById('aiChatMessages');
        if (messagesContainer) {
            setTimeout(() => {
                messagesContainer.scrollTop = messagesContainer.scrollHeight;
            }, 100);
        }
    }
}

// Глобальные функции для совместимости с шаблоном
function toggleAIChat() {
    if (window.aiChat) {
        window.aiChat.toggle();
    }
}

function closeAIChat() {
    if (window.aiChat) {
        window.aiChat.close();
    }
}

function sendAIMessage() {
    if (window.aiChat) {
        window.aiChat.sendMessage();
    }
}

function sendQuickMessage(text) {
    if (window.aiChat) {
        window.aiChat.sendQuickMessage(text);
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    // Проверяем, есть ли AI виджет на странице
    if (document.getElementById('aiChatWidget')) {
        window.aiChat = new AIChat();
        
        // Добавляем стили для предложений
        const style = document.createElement('style');
        style.textContent = `
            .ai-suggestions {
                background: var(--muted);
                border-radius: var(--radius);
                padding: 12px;
                margin-top: 8px;
            }
            
            .ai-suggestions-title {
                font-size: 12px;
                font-weight: 600;
                color: var(--muted-foreground);
                margin-bottom: 8px;
            }
            
            .ai-suggestions-buttons {
                display: flex;
                flex-direction: column;
                gap: 6px;
            }
            
            .ai-suggestion-btn {
                background: var(--card);
                border: 1px solid var(--border);
                color: var(--card-foreground);
                padding: 8px 12px;
                border-radius: 6px;
                font-size: 13px;
                cursor: pointer;
                transition: all 0.2s;
                text-align: left;
            }
            
            .ai-suggestion-btn:hover {
                background: var(--accent);
                color: var(--accent-foreground);
                border-color: var(--accent);
            }
            
            .suggestions-message {
                margin-bottom: 15px;
            }
            
            .typing-message {
                margin-bottom: 15px;
            }
        `;
        document.head.appendChild(style);
    }
});

// Экспорт для использования в других модулях
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AIChat;
}