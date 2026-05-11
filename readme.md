Архитектура Full local-first + Bring Your Own Backup
Все содержательные данные хранятся только на устройстве пользователя
Сервер приложения не хранит пользовательские реестры и не получает маппинг ID → название/человек/заметка
Сервер хранит и синхронизирует: айдишники и их счетчики для того что бы можно было собирать общие счетчики (руководитель и подчиненные)
Резервное копирование выполняет сам пользователь


Приложение не инициирует и не требует внесения персональных данных третьих лиц. Если пользователь использует приложение для учёта людей, содержательные данные остаются локально на его устройстве. Поэтому оператором таких данных выступает сам пользователь/организация, а сервис не хранит эти данные на своей инфраструктуре.

не передавать пользовательские данные в аналитику, crash reporting и поддержку;

Разные типы локальных заметок номера / почта / теги и тп

случайные локальные UUID для сущностей ?? - мб все же на сервере что бы коллизий небыло?
локальная БД: SQLite/IndexedDB; ??? и что там с шифрованием
пароль/биометрия для доступа к приложению;
экспорт/импорт зашифрованного бэкапа;

Сервер хранит названия статусов, но только для агрегированной статистики:

{
"status_id": "s1",
"status_name": "Активный",
"count": 12
}

SPA → PWA → пользователь открывает ссылку → “Add to Home Screen” / “Установить приложение”

Стек:

React + TypeScript
Vite
MUI
React Router
React Hook Form
Zod

PWA:
vite-plugin-pwa + Workbox

Нужно:

manifest.webmanifest
service worker
offline cache app shell
cache busting при обновлениях
install prompt для Android
инструкция “Add to Home Screen” для iOS

Типовая структура:

src/
    app/
    db/
    sync/
    features/
public/
    manifest.webmanifest
    icons/
service-worker.ts


Можно запросить persistent storage:

if (navigator.storage?.persist) {
const granted = await navigator.storage.persist();
console.log("Persistent storage:", granted);
}


Экспорт/импорт
Формат backup-файла

Лучший практичный вариант:

.myapp-backup.zip

Внутри:

manifest.json
data.json
attachments/
checksums.json

Пример:

{
"app": "my-app",
"schemaVersion": 4,
"createdAt": "2026-05-11T12:00:00.000Z",
"device": "browser",
"tables": {
"lists": 12,
"items": 340,
"forms": 8
}
}

Для ZIP:

fflate

Для валидации:

Zod

Для контрольных сумм:

Web Crypto API / SHA-256