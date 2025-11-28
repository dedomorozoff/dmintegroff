// Common JavaScript functions

// Copy text to clipboard
function copyToClipboard(text) {
    const fullUrl = window.location.origin + text;
    navigator.clipboard.writeText(fullUrl).then(() => {
        alert('URL скопирован: ' + fullUrl);
    }).catch(err => {
        console.error('Failed to copy:', err);
    });
}

// Confirm before delete
function confirmDelete(message) {
    return confirm(message || 'Вы уверены, что хотите удалить?');
}

// Generate mapping config from form
function generateMapping(event) {
    event.preventDefault();
    
    const mapping = {};
    const form = document.getElementById('mappingForm');
    const inputs = form.querySelectorAll('input[type="text"]');
    
    inputs.forEach(input => {
        const fieldName = input.name.replace('field_', '');
        const newName = input.value.trim();
        const ignoreCheckbox = form.querySelector(`input[name="ignore_${fieldName}"]`);
        
        if (!ignoreCheckbox.checked && newName) {
            mapping[newName] = fieldName;
        }
    });
    
    document.getElementById('mappingConfig').value = JSON.stringify(mapping);
    form.submit();
}
