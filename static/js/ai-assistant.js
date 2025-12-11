/**
 * AI Assistant Chat Widget для dmIntegroff
 * Простой чат-помощник в стиле современных сайтов
 */

class AIAssistant {
    constructor() {
        this.isOpen = false;
        this.isTyping = false;
        this.chatHistory = [];
        this.integrationContext = null;
        
        this.initializeElements();
        this.bindEvents();
        this.checkAIStatus();
    }

    initializeElements() {
        this.widget = document.getElementById('aiChatWidget');
        this.toggle = document.getElementById('aiChatToggle');
        this.window = document.getElementById('aiChatWindow');
        this.messages = document.getElementById('aiChatMessages');
        this.input = document.getElementById('aiChatInput');
        this.sendButton = document.getElementById('aiSendButton');
        this.status = document.getElementById('aiChatStatus');
        this.badge = document.getElementById('aiChatBadge');
    }

    bindEvents() {
        // Enter для отправки
        this.input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Клик вне окна для закрытия
        document.addEventListener('click', (e) => {
            if (this.isOpen && !this.widget.contains(e.target)) {
                this.closeChat();
            }
        });
    }

    async checkAIStatus() {
        try {
            const response = await fetch('/api/ai/status');
            const status = await response.json();
            
            if (status.configured && status.available) {
                this.badge.textContent = 'AI';
                this.badge.style.background = '#2ed573'; // зеленый
            } else {
                this.badge.textContent = '!';
                this.badge.style.background = '#ff4757'; // красный
            }
        } catch (error) {
            this.badge.textContent = '?';
            this.badge.style.background = '#ffa502'; // оранжевый
        }
    }

    toggleChat() {
        if (this.isOpen) {
            this.closeChat();
        } else {
            this.openChat();
        }
    }

    openChat() {
        this.isOpen = true;
        this.window.classList.add('open');
        this.input.focus();
        
        // Скрываем badge когда открыт
        this.badge.style.display = 'none';
    }

    closeChat() {
        this.isOpen = false;
        this.window.classList.remove('open');
        
        // Показываем badge когда закрыт
        this.badge.style.display = 'block';
    }

    async sendMessage() {
        const message = this.input.value.trim();
        if (!message || this.isTyping) return;

        // Добавляем сообщение пользователя
        this.addMessage(message, 'user');
        this.input.value = '';

        // Показываем индикатор печати
        this.showTyping();

        try {
            const response = await fetch('/api/ai/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    message: message,
                    history: this.chatHistory,
                    context: this.integrationContext,
                    sample_data: this.getSampleData()
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            const result = await response.json();
            
            this.hideTyping();
            this.addMessage(result.response, 'ai');
            
            // Обновляем историю
            this.chatHistory.push(
                { role: 'user', content: message },
                { role: 'assistant', content: result.response }
            );

        } catch (error) {
            console.error('Error sending message:', error);
            this.hideTyping();
            this.addMessage('Извините, произошла ошибка. Попробуйте еще раз.', 'ai');
        }
    }

    sendQuickMessage(message) {
        this.input.value = message;
        this.sendMessage();
    }

    addMessage(content, sender) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `${sender}-message`;
        
        const now = new Date().toLocaleTimeString('ru-RU', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
        
        const avatar = sender === 'user' ? '👤' : '🤖';
        
        messageDiv.innerHTML = `
            <div class="${sender}-avatar">${avatar}</div>
            <div class="${sender}-message-content">
                <div class="${sender}-message-text">${this.formatMessage(content)}</div>
                <div class="${sender}-message-time">${now}</div>
            </div>
        `;
        
        this.messages.appendChild(messageDiv);
        this.scrollToBottom();
    }

    formatMessage(content) {
        return content
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/g, '<em>$1</em>')
            .replace(/`(.*?)`/g, '<code style="background:#f1f3f4;padding:2px 4px;border-radius:3px;">$1</code>')
            .replace(/\n/g, '<br>');
    }

    showTyping() {
        this.isTyping = true;
        this.sendButton.disabled = true;
        
        const typingDiv = document.createElement('div');
        typingDiv.className = 'ai-message';
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
        
        this.messages.appendChild(typingDiv);
        this.scrollToBottom();
    }

    hideTyping() {
        this.isTyping = false;
        this.sendButton.disabled = false;
        
        const typingIndicator = document.getElementById('typingIndicator');
        if (typingIndicator) {
            typingIndicator.remove();
        }
    }

    scrollToBottom() {
        this.messages.scrollTop = this.messages.scrollHeight;
    }

    getSampleData() {
        if (window.samplePayload) {
            try {
                return JSON.parse(window.samplePayload);
            } catch (e) {
                return {};
            }
        }
        return {};
    }

    setIntegrationContext(context) {
        this.integrationContext = context;
    }
}

// Глобальные функции
window.toggleAIChat = function() {
    if (window.aiAssistant) {
        window.aiAssistant.toggleChat();
    }
};

window.closeAIChat = function() {
    if (window.aiAssistant) {
        window.aiAssistant.closeChat();
    }
};

window.sendAIMessage = function() {
    if (window.aiAssistant) {
        window.aiAssistant.sendMessage();
    }
};

window.sendQuickMessage = function(message) {
    if (window.aiAssistant) {
        window.aiAssistant.sendQuickMessage(message);
    }
};

// Обновляем глобальную функцию openAIAssistant
window.openAIAssistant = function(context = null) {
    if (window.aiAssistant) {
        if (context) {
            window.aiAssistant.setIntegrationContext(context);
        }
        window.aiAssistant.openChat();
    } else {
        // Показываем сообщение если AI не доступен
        alert('AI Ассистент доступен на страницах с интеграциями');
    }
};

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    // Проверяем есть ли AI виджет на странице
    if (document.getElementById('aiChatWidget')) {
        window.aiAssistant = new AIAssistant();
    }
});

