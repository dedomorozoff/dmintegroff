/**
 * JSON Syntax Highlighter
 * Простая подсветка синтаксиса JSON для textarea
 */

class JSONHighlighter {
    constructor(textareaId, previewId) {
        this.textarea = document.getElementById(textareaId);
        this.preview = document.getElementById(previewId);
        
        if (!this.textarea) {
            console.error(`Textarea with id "${textareaId}" not found`);
            return;
        }
        
        // Проверяем, не инициализирован ли уже
        if (this.textarea.dataset.highlighterInitialized === 'true') {
            console.warn(`JSONHighlighter: Already initialized for ${textareaId}`);
            return;
        }
        
        this.init();
    }
    
    init() {
        console.log('JSONHighlighter: Initializing for', this.textarea.id);
        
        // Ищем существующий preview элемент или создаем новый
        if (!this.preview) {
            const existingPreview = document.getElementById(this.textarea.id + '_preview');
            if (existingPreview) {
                this.preview = existingPreview;
                console.log('JSONHighlighter: Using existing preview element', this.preview.id);
            } else {
                this.preview = document.createElement('div');
                this.preview.id = this.textarea.id + '_preview';
                this.preview.className = 'json-preview';
                this.textarea.parentNode.insertBefore(this.preview, this.textarea);
                console.log('JSONHighlighter: Created new preview element', this.preview.id);
            }
        }
        
        // Стилизуем элементы
        this.setupStyles();
        
        // Обработчики событий
        this.textarea.addEventListener('input', () => this.update());
        this.textarea.addEventListener('scroll', () => this.syncScroll());
        
        // Первоначальное обновление
        this.update();
        
        // Помечаем как инициализированный
        this.textarea.dataset.highlighterInitialized = 'true';
        
        console.log('JSONHighlighter: Initialized successfully');
    }
    
    setupStyles() {
        // Получаем родительский контейнер
        const container = this.textarea.parentElement;
        container.style.position = 'relative';
        
        // Настраиваем preview (СНАЧАЛА!)
        this.preview.style.position = 'absolute';
        this.preview.style.top = '0';
        this.preview.style.left = '0';
        this.preview.style.width = '100%';
        this.preview.style.height = '100%';
        this.preview.style.padding = '1rem';
        this.preview.style.fontFamily = "'Courier New', monospace";
        this.preview.style.fontSize = '0.95em';
        this.preview.style.lineHeight = '1.5';
        this.preview.style.overflow = 'hidden'; // Скролл только у textarea
        this.preview.style.background = '#1a1f2e';
        this.preview.style.borderRadius = '8px';
        this.preview.style.border = '1px solid #2d3748';
        this.preview.style.pointerEvents = 'none';
        this.preview.style.setProperty('z-index', '1', 'important');
        this.preview.style.whiteSpace = 'pre-wrap'; // Изменено с pre на pre-wrap
        this.preview.style.wordWrap = 'break-word';
        this.preview.style.boxSizing = 'border-box';
        this.preview.style.setProperty('visibility', 'visible', 'important');
        this.preview.style.setProperty('display', 'block', 'important');
        
        // Настраиваем textarea (ПОТОМ!)
        this.textarea.style.position = 'relative';
        this.textarea.style.setProperty('color', 'transparent', 'important');
        this.textarea.style.setProperty('caret-color', '#abb2bf', 'important');
        this.textarea.style.setProperty('background', 'transparent', 'important');
        this.textarea.style.setProperty('z-index', '2', 'important');
        
        console.log('Styles applied:', {
            textareaColor: this.textarea.style.color,
            textareaZIndex: this.textarea.style.zIndex,
            previewZIndex: this.preview.style.zIndex,
            previewPosition: this.preview.style.position
        });
    }
    
    update() {
        const text = this.textarea.value;
        const highlighted = this.highlight(text);
        this.preview.innerHTML = highlighted;
    }
    
    syncScroll() {
        this.preview.scrollTop = this.textarea.scrollTop;
        this.preview.scrollLeft = this.textarea.scrollLeft;
    }
    
    highlight(text) {
        return JSONHighlighter.highlightText(text);
    }
}

// Статический метод для подсветки JSON без textarea
JSONHighlighter.highlightText = function(text) {
    if (!text) return '';
    
    // Экранируем HTML
    text = text.replace(/&/g, '&amp;')
               .replace(/</g, '&lt;')
               .replace(/>/g, '&gt;');
    
    // Подсветка различных элементов JSON (с inline стилями для надежности)
    text = text
        // Строки (включая плейсхолдеры) - ЯРКИЙ ЗЕЛЕНЫЙ
        .replace(/"([^"\\]*(\\.[^"\\]*)*)"/g, (match) => {
            // Проверяем, содержит ли строка плейсхолдер
            if (match.includes('{{') && match.includes('}}')) {
                return match.replace(/\{\{([^}]+)\}\}/g, 
                    '<span style="color: #61afef; font-weight: bold; background: rgba(97, 175, 239, 0.15);">{{$1}}</span>');
            }
            return `<span style="color: #98c379;">${match}</span>`;
        })
        // Числа - ОРАНЖЕВЫЙ
        .replace(/\b(-?\d+\.?\d*)\b/g, '<span style="color: #d19a66;">$1</span>')
        // Булевы значения - ГОЛУБОЙ
        .replace(/\b(true|false)\b/g, '<span style="color: #56b6c2;">$1</span>')
        // null - ФИОЛЕТОВЫЙ
        .replace(/\bnull\b/g, '<span style="color: #c678dd;">null</span>')
        // Ключи (слова перед двоеточием) - КРАСНЫЙ
        .replace(/("[\w\s_-]+")\s*:/g, '<span style="color: #e06c75; font-weight: 500;">$1</span>:')
        // Плейсхолдеры вне строк - СИНИЙ
        .replace(/\{\{([^}]+)\}\}/g, '<span style="color: #61afef; font-weight: bold; background: rgba(97, 175, 239, 0.15);">{{$1}}</span>');
    
    return text;
};

// Функция для подсветки статического JSON в <pre> элементах
function highlightStaticJSON(elementId) {
    const element = document.getElementById(elementId);
    if (!element) {
        console.warn('Element not found:', elementId);
        return;
    }
    
    const text = element.textContent;
    const highlighted = JSONHighlighter.highlightText(text);
    element.innerHTML = highlighted;
    console.log('Static JSON highlighted:', elementId);
}

// Автоматическая инициализация отключена
// Инициализация происходит вручную в шаблоне страницы

// Экспортируем функцию подсветки глобально
window.highlightJSON = JSONHighlighter.highlightText;
