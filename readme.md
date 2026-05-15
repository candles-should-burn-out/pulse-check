# Pulse Check

Pulse Check - privacy-oriented local-first приложение для учета сущностей, статусов и агрегированной статистики. Пользовательские названия, заметки, поля, атрибуты и маппинги по умолчанию остаются на устройстве пользователя; сервер хранит только технические идентификаторы, статусы, счетчики и служебные данные, необходимые для работы сервиса.

## Документация

- [Архитектура в формате arc42](docs/architecture/arc42.md)
- [Architecture Decision Records](docs/adr/readme.md)
- [Backend](backend/readme.md)
- [Frontend](frontend/readme.md)

### Markdown style

В Markdown-файлах не делаем ручной перенос строк только из-за достижения лимита длины строки. В среде разработки включен визуальный автоперенос, поэтому абзацы пишем одной строкой; переносы оставляем для смысловой структуры: новые абзацы, списки, таблицы и fenced code blocks.

## Локальная разработка

Если нужны локальные секреты или переопределения, скопируйте `.env.example` в `.env`. Файл `.env` не коммитится.

Запуск полного локального стенда:

```sh
task up
```

Перезапуск только application-сервисов:

```sh
task backend-restart
task frontend-restart
```

Полезные локальные URL:

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Keycloak Admin Console: `http://localhost:8081/admin`
- Grafana: `http://localhost:3001`

Локальный тестовый пользователь: `admin` / `admin`

## Проверки

```sh
cd backend && go test ./...
cd frontend && npm run lint
cd frontend && npm run build
```

Ручные проверки:

- `/` открывается без авторизации.
- `/app` редиректит на Keycloak.
- После входа frontend отправляет `Authorization: Bearer <token>` и защищенные API-вызовы работают.
- Logout возвращает пользователя на публичную страницу.
