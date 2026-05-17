# ADR-0003: Server stores only technical aggregates

Status: Accepted

## Context

Pulse Check нужна серверная координация и агрегированная статистика, но privacy goal из ADR-0001 требует не хранить на сервере пользовательские персональные данные и персональные маппинги.

Некоторые совместные сценарии требуют стабильных технических идентификаторов и агрегированных счетчиков, чтобы владелец и участники видели сопоставимую статистику без раскрытия конкретных сущностей.

## Decision

Сервер хранит только technical identifiers, status identifiers/names там, где это нужно для общей статистики, и aggregate counters.

Разрешенные категории данных:

- technical identifiers, например `entity_id`, `state_id`, `client_id`, `status_id`;
- status identifiers, names и presentation metadata, если это явно разрешено отдельным ADR;
- aggregate counters, например `count`;
- технические timestamps и ownership identifiers, необходимые для доступа и синхронизации.

По умолчанию сервер не должен хранить attributes, notes, user fields или `ID -> name/person/note` mappings.

ADR-0005 является осознанным уточнением этой границы: сервер хранит имена и цвета статусов, потому что они нужны для совместного сценария работы и агрегированной статистики.

## Данные и База Данных

Не затрагивается.

## API

Не затрагивается.

## Ограничения

- Новые backend tables и API не должны принимать произвольные пользовательские заметки, атрибуты или персональные маппинги без отдельного ADR.
- Logs, traces, metrics и support workflows считаются частью той же data minimization boundary.
- Readable names допускаются на сервере только как явное исключение, описанное отдельным ADR.
- Агрегаты не должны позволять восстановить конкретную сущность или персональный профиль пользователя.

## Consequences

- Backend APIs остаются согласованными с privacy boundary.
- Product features, которым нужны readable names или notes, должны использовать local data или требовать отдельного explicit architecture decision.
- Aggregated owner/participant views можно строить вокруг statuses и counts без раскрытия конкретных entities.
- Developers должны проектировать observability и support flows как потенциальный канал утечки пользовательских данных.

## Проработка

- Для каждого нового shared feature проверять, не требует ли он отдельного ADR об исключении из privacy boundary.
- Спроектировать правила redaction для будущих frontend error events и support diagnostics.
