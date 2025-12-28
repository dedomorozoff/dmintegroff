// Популярные шаблоны систем для интеграций
let popularTemplates = {};

// Загрузка шаблонов из файлов
async function loadTemplatesFromFiles() {
    const categories = ['crm', 'messengers', 'email', 'analytics', 'payments', 'databases', 'other'];
    
    for (const category of categories) {
        try {
            // Пытаемся получить список файлов в категории
            const templateFiles = {
                crm: ['salesforce.json', 'hubspot.json', 'pipedrive.json', 'amocrm.json', 'bitrix24.json'],
                messengers: ['telegram.json', 'slack.json', 'discord.json', 'whatsapp.json', 'vk.json'],
                email: ['mailchimp.json', 'sendgrid.json', 'mailgun.json'],
                analytics: ['google-analytics.json', 'mixpanel.json', 'yandex-metrica.json'],
                payments: ['stripe.json', 'paypal.json', 'yookassa.json', 'robokassa.json'],
                databases: ['mysql.json', 'postgresql.json', 'mongodb.json'],
                other: ['jira.json', 'trello.json', 'zapier.json', 'google-sheets.json', 'airtable.json']
            };
            
            const files = templateFiles[category] || [];
            
            for (const file of files) {
                try {
                    const templateResponse = await fetch(`/integration-templates/${category}/${file}`);
                    if (templateResponse.ok) {
                        const templateData = await templateResponse.json();
                        const key = file.replace('.json', '');
                        popularTemplates[key] = {
                            type: templateData.type,
                            template: typeof templateData.template === 'string' 
                                ? templateData.template 
                                : JSON.stringify(templateData.template, null, 2),
                            description: templateData.description,
                            category: templateData.category,
                            name: templateData.name
                        };
                        console.log(`Загружен шаблон: ${key} (${templateData.name})`);
                    } else {
                        console.warn(`Не удалось загрузить шаблон ${file}: HTTP ${templateResponse.status}`);
                    }
                } catch (fileError) {
                    console.warn(`Ошибка при загрузке шаблона ${file}:`, fileError);
                }
            }
        } catch (error) {
            console.warn(`Не удалось загрузить шаблоны категории ${category}:`, error);
        }
    }
}

// Популярные шаблоны будут загружены только из файлов

// Инициализация шаблонов
async function initializeTemplates() {
    try {
        // Загружаем шаблоны только из файлов
        await loadTemplatesFromFiles();
        
        if (Object.keys(popularTemplates).length === 0) {
            console.warn('Не удалось загрузить шаблоны из файлов. Проверьте доступность /integration-templates/');
        } else {
            console.log('Шаблоны успешно загружены:', Object.keys(popularTemplates).length);
            console.log('Доступные шаблоны:', Object.keys(popularTemplates));
        }
    } catch (error) {
        console.error('Ошибка при загрузке шаблонов:', error);
    }
    
    // Обновляем селект с шаблонами
    updateTemplateSelect();
}

// Обновление селекта с шаблонами
function updateTemplateSelect() {
    const select = document.getElementById('popularTemplates');
    if (!select) return;
    
    // Очищаем текущие опции (кроме первой)
    while (select.children.length > 1) {
        select.removeChild(select.lastChild);
    }
    
    // Если шаблоны не загружены, показываем сообщение
    if (Object.keys(popularTemplates).length === 0) {
        const option = document.createElement('option');
        option.value = '';
        option.textContent = 'Шаблоны недоступны (перезапустите сервер)';
        option.disabled = true;
        select.appendChild(option);
        return;
    }
    
    // Группируем шаблоны по категориям
    const categories = {
        crm: 'CRM системы',
        messengers: 'Мессенджеры',
        email: 'Email сервисы',
        analytics: 'Аналитика',
        payments: 'Платежи',
        databases: 'Базы данных',
        other: 'Другие'
    };
    
    const templatesByCategory = {};
    
    Object.entries(popularTemplates).forEach(([key, template]) => {
        const category = template.category || 'other';
        if (!templatesByCategory[category]) {
            templatesByCategory[category] = [];
        }
        templatesByCategory[category].push({ key, template });
    });
    
    // Добавляем опции по категориям
    Object.entries(categories).forEach(([categoryKey, categoryName]) => {
        if (templatesByCategory[categoryKey] && templatesByCategory[categoryKey].length > 0) {
            const optgroup = document.createElement('optgroup');
            optgroup.label = categoryName;
            
            templatesByCategory[categoryKey].forEach(({ key, template }) => {
                const option = document.createElement('option');
                option.value = key;
                option.textContent = template.name || key;
                optgroup.appendChild(option);
            });
            
            select.appendChild(optgroup);
        }
    });
}

// Загрузка популярного шаблона
function loadPopularTemplate() {
    const select = document.getElementById('popularTemplates');
    const templateKey = select.value;
    const loadBtn = document.getElementById('loadTemplateBtn');
    
    // Обновляем состояние кнопки
    loadBtn.disabled = !templateKey;
    
    if (!templateKey) return;
    
    const template = popularTemplates[templateKey];
    if (!template) {
        if (typeof window.showNotification === 'function') {
            window.showNotification('Шаблон не найден', 'error');
        }
        return;
    }
    
    // Ищем textarea для шаблона (может быть outputTemplate или output_template)
    let textarea = document.getElementById('outputTemplate');
    if (!textarea) {
        textarea = document.getElementById('output_template');
    }
    
    if (!textarea) {
        if (typeof window.showNotification === 'function') {
            window.showNotification('Не найдено поле для шаблона', 'error');
        }
        return;
    }
    
    // Подтверждение замены
    if (textarea.value.trim() && !confirm('Заменить текущий шаблон? Несохраненные изменения будут потеряны.')) {
        select.value = '';
        loadBtn.disabled = true;
        return;
    }
    
    // Устанавливаем тип шаблона (может быть templateType или template_type)
    let templateTypeSelect = document.getElementById('templateType');
    if (!templateTypeSelect) {
        templateTypeSelect = document.getElementById('template_type');
    }
    
    if (templateTypeSelect) {
        templateTypeSelect.value = template.type;
    }
    
    // Загружаем шаблон
    textarea.value = template.template;
    textarea.dispatchEvent(new Event('input'));
    
    // Обновляем редактор
    if (typeof updateTemplateEditor === 'function') {
        updateTemplateEditor();
    }
    
    // Обновляем подсветку синтаксиса если доступна
    if (typeof JSONHighlighter !== 'undefined') {
        const previewId = textarea.id + '_preview';
        const preview = document.getElementById(previewId);
        if (preview && typeof highlightJSON !== 'undefined') {
            preview.innerHTML = highlightJSON(textarea.value);
        }
    }
    
    // Показываем уведомление
    const templateName = select.options[select.selectedIndex].text;
    if (typeof window.showNotification === 'function') {
        window.showNotification(`✅ Шаблон "${templateName}" загружен!\n${template.description}`, 'success');
    }
    
    // Сбрасываем выбор
    select.value = '';
    loadBtn.disabled = true;
    
    // Обновляем иконки
    if (window.lucide) {
        window.lucide.createIcons();
    }
}

// Обработчик изменения выбора популярных шаблонов
document.addEventListener('change', (e) => {
    if (e.target.id === 'popularTemplates') {
        const loadBtn = document.getElementById('loadTemplateBtn');
        if (loadBtn) {
            loadBtn.disabled = !e.target.value;
        }
        
        // Обновляем иконки
        if (window.lucide) {
            window.lucide.createIcons();
        }
    }
});

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', initializeTemplates);