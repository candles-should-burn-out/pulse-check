# Архитектура

- Full Local-First + Bring Your Own Backup
- Атрибуты могут быть разных типов: номер / почта / тег / заметка
- Для сущностей используются UUID (генерируются на стороне сервера)
  - В офлайн режиме действия превращаются в очередь эвентов ждущих подтверждение учета от сервера
- Можно создавать подчиненных (тех кто будет помогать собирать статистику) - они получают информацию о используемых статусах и их статистика отображается у руководителя (обобщенная по статусам, без конкретных сущностей)
- Возможность экспорта / импорта зашифрованного бэкапа
- Резервное копирование выполняет сам пользователь

## Обработка персональных данных

- Все пользовательские данные хранятся только на устройстве пользователя
- Сервер не получает и не хранит: Атрибуты / маппинг `ID → название / человек / заметка` / персональные данные третьих лиц
- Запрещено передавать пользовательские данные в: аналитику / crash reporting / поддержку / внешние сервисы без явного действия пользователя

## Frontend

- TODO: Доступ к приложению можно защищать паролем или биометрией
- TODO: Данные локально можно шифровать
- SPA / PWA: пользователь открывает ссылку → "Add to Home Screen" / "Установить приложение"
- Стек: React / TypeScript / Vite / MUI / React Router / React Hook Form / IndexedDB

### PWA

- vite-plugin-pwa, Workbox
- manifest.webmanifest
- service worker
- offline cache app shell
- cache busting при обновлениях
- install prompt для Android
- инструкция “Add to Home Screen” для iOS

### Типовая структура:

```
src/
    app/
    db/
    sync/
    features/
public/
    manifest.webmanifest
    icons/
service-worker.ts
```

### Persistent storage

```
if (navigator.storage?.persist) {
    const granted = await navigator.storage.persist();
    console.log("Persistent storage:", granted);
}
```

### Backup-файл

- Для ZIP: fflate
- Для валидации: Zod
- Для контрольных сумм: Web Crypto API / SHA-256
- Содержимое: manifest.json / data.json / attachments / checksums.json

## Backend

- Сервер хранит только идентификаторы сущностей (entity_id), статусы (id, name) и агрегированные счётчики (client_id, status_id, count)