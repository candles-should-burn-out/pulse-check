- В шаблонах статусов сделать более насыщенные цвета  
- Бэкапы
  - Спроектировать точный формат `manifest.json`, `data.json` и `checksums.json`.
  - Выбрать схему шифрования backup и UX хранения ключа или пароля.
  - Спроектировать conflict resolution для импорта backup поверх существующих локальных данных.
- Планируемые PWA-возможности: `manifest.webmanifest`, service worker, Workbox cache, offline app shell, cache busting при обновлениях, install prompt для Android и инструкция "Add to Home Screen" для iOS.
- Планируемое локальное хранение: IndexedDB, опциональное локальное шифрование и `navigator.storage.persist()` там, где браузер это поддерживает.
- Переделать localhost: Frontend экспортирует browser fetch traces в `http://localhost:4318`, когда tracing включен.
- Keycloak realm drift возможен после первого import, потому что существующий realm не перезатирается последующими `--import-realm` запусками.
- Ошибка production URL при первом realm import может сохраниться в базе Keycloak и потребовать ручного исправления через Admin Console.
- Access token остается валидным до expiration после disabled user.
- Browser storage может быть очищен; persistent storage повышает шансы, но не является backup guarantee.
- Если пользователь не настроил бэкапы показывать предупреждение: Статистика сохранится, а вот детали нет!
- Но важный нюанс: если этот же keycloak/pulse-check-realm.json когда-нибудь использовать для production-импорта, тогда sslRequired: "none" будет слишком мягкой настройкой. Перед production лучше сделать отдельный realm export/seed или production override, где будет HTTPS-режим и реальные redirect URI вроде https://your-domain/app/*.
- Хороший будущий шаг: разделить keycloak/pulse-check-realm.local.json и production seed/config, чтобы локальная удобная настройка не могла случайно уехать в боевое окружение.

## Offline and sync flow:

1. Пользователь выполняет действия offline.
2. Frontend сохраняет действия локально как queued events.
3. Когда сеть и авторизация доступны, frontend отправляет события на сервер для acknowledgement.
4. Conflict resolution и retry policy пока TBD.

## Backup/import flow:

1. Пользователь явно запускает export backup.
2. Frontend создает зашифрованный архив с `manifest.json`, `data.json`, опциональным `attachments/` и `checksums.json`.
3. Для ZIP планируется `fflate`, для валидации - Zod, для checksums - Web Crypto API/SHA-256.
4. Пользователь самостоятельно хранит backup и позже явно импортирует его.