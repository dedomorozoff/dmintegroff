# 🛠 dmIntegroff CLI - Инструмент администрирования

## Установка

```bash
# Сборка CLI инструмента
go build -o dmIntegroff-admin cmd/admin/main.go

# Или для Windows
go build -o dmIntegroff-admin.exe cmd/admin/main.go
```

## Использование

### Интерактивный режим

```bash
./dmIntegroff-admin
```

Меню:
```
=== dmIntegroff Admin CLI ===
1. Выполнить миграцию базы данных
2. Сбросить пароль администратора
3. Сгенерировать случайный URL
0. Выход
```

### Командная строка

#### 1. Миграция базы данных

```bash
./dmIntegroff-admin --migrate
```

**Что делает:**
- Создает все таблицы (users, integrations, request_logs)
- Создает индексы для оптимизации запросов
- Создает администратора по умолчанию (admin/admin)

**Когда использовать:**
- При первом развертывании
- После обновления структуры БД
- При переходе на новую БД

**Пример вывода:**
```
=== Миграция базы данных ===
Тип базы данных: sqlite

📊 Выполнение миграций...
✅ Таблицы созданы/обновлены

👤 Создание администратора по умолчанию...

🎉 Миграция завершена успешно!

📝 Учетные данные по умолчанию:
   Логин: admin
   Пароль: admin

⚠️  ВАЖНО: Смените пароль после первого входа!
```

#### 2. Сброс пароля администратора

```bash
./dmIntegroff-admin --reset-password
```

**Что делает:**
- Находит первого администратора в БД
- Запрашивает новый пароль (скрытый ввод)
- Хеширует пароль через bcrypt
- Сохраняет в БД

**Когда использовать:**
- Забыли пароль администратора
- Нужно сменить пароль по соображениям безопасности
- После компрометации учетных данных

**Пример использования:**
```bash
$ ./dmIntegroff-admin --reset-password

=== Сброс пароля администратора ===
Найден администратор: admin
Введите новый пароль: ********
✅ Пароль успешно изменен!
```

#### 3. Генерация случайного URL

```bash
./dmIntegroff-admin --generate-secret
```

**Что делает:**
- Генерирует случайный 32-символьный путь
- Показывает инструкции по настройке

**Когда использовать:**
- Для скрытия админ-панели от сканеров
- Дополнительная безопасность через obscurity
- Защита от автоматических атак

**Пример вывода:**
```
=== Генерация случайного URL ===

🔐 Случайный путь для приложения:
   /a3f9c2e1b4d8f7a6c5e2d9b8a7f6e5d4

Добавьте в .env файл:
   APP_PATH=/a3f9c2e1b4d8f7a6c5e2d9b8a7f6e5d4

После этого приложение будет доступно по адресу:
   http://localhost:8080/a3f9c2e1b4d8f7a6c5e2d9b8a7f6e5d4

⚠️  Не забудьте перезапустить сервер!
```

## Примеры использования

### Первоначальная настройка

```bash
# 1. Выполните миграцию
./dmIntegroff-admin --migrate

# 2. Смените пароль администратора
./dmIntegroff-admin --reset-password

# 3. (Опционально) Сгенерируйте случайный URL
./dmIntegroff-admin --generate-secret

# 4. Запустите сервер
./dmIntegroff
```

### Восстановление доступа

```bash
# Если забыли пароль
./dmIntegroff-admin --reset-password
```

### Миграция на новый сервер

```bash
# 1. Скопируйте .env файл
cp /old/server/.env .

# 2. Выполните миграцию
./dmIntegroff-admin --migrate

# 3. Восстановите данные из бэкапа (если есть)
mysql -u user -p dmIntegroff < backup.sql
```

## Переменные окружения

CLI использует те же переменные окружения, что и основное приложение:

```env
# База данных
DB_TYPE=sqlite          # или mysql
DB_DSN=dmIntegroff.db      # или строка подключения MySQL

# Для MySQL
DB_DSN=user:password@tcp(localhost:3306)/dmIntegroff?charset=utf8mb4&parseTime=True&loc=Local
```

## Безопасность

### Рекомендации:

1. **Храните CLI в безопасном месте**
   ```bash
   chmod 700 dmIntegroff-admin
   chown root:root dmIntegroff-admin
   ```

2. **Не коммитьте в git**
   - CLI уже в .gitignore
   - Собирайте на сервере

3. **Используйте сильные пароли**
   - Минимум 12 символов
   - Буквы, цифры, спецсимволы
   - Не используйте словарные слова

4. **Регулярно меняйте пароли**
   ```bash
   # Каждые 90 дней
   ./dmIntegroff-admin --reset-password
   ```

## Troubleshooting

### Ошибка: "Database connection failed"

**Проблема:** Не удается подключиться к БД

**Решение:**
```bash
# Проверьте .env файл
cat .env | grep DB_

# Проверьте доступность MySQL
mysql -h localhost -u user -p

# Проверьте права доступа
ls -la dmIntegroff.db  # для SQLite
```

### Ошибка: "Admin user not found"

**Проблема:** Нет администратора в БД

**Решение:**
```bash
# Выполните миграцию заново
./dmIntegroff-admin --migrate
```

### Ошибка: "Permission denied"

**Проблема:** Нет прав на выполнение

**Решение:**
```bash
# Добавьте права на выполнение
chmod +x dmIntegroff-admin
```

## Автоматизация

### Cron задачи

```bash
# Автоматическая смена пароля каждые 90 дней
0 0 1 */3 * /opt/dmIntegroff/scripts/rotate-password.sh

# Скрипт rotate-password.sh:
#!/bin/bash
NEW_PASSWORD=$(openssl rand -base64 16)
echo "$NEW_PASSWORD" | /opt/dmIntegroff/dmIntegroff-admin --reset-password
echo "New password: $NEW_PASSWORD" | mail -s "dmIntegroff Password Rotated" admin@example.com
```

### CI/CD интеграция

```yaml
# .github/workflows/deploy.yml
- name: Run migrations
  run: |
    ./dmIntegroff-admin --migrate
```

## Дополнительные команды (планируется)

- [ ] `--backup` - Создание резервной копии БД
- [ ] `--restore` - Восстановление из резервной копии
- [ ] `--create-user` - Создание нового пользователя
- [ ] `--list-users` - Список всех пользователей
- [ ] `--delete-user` - Удаление пользователя
- [ ] `--stats` - Статистика использования
- [ ] `--cleanup-logs` - Очистка старых логов

## Поддержка

При возникновении проблем:
1. Проверьте логи: `tail -f dmIntegroff.log`
2. Проверьте подключение к БД
3. Создайте issue на GitHub

---

**Версия:** 1.1.0  
**Последнее обновление:** 2024-11-29
