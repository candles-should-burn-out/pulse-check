# ADR-0003: Server stores only technical aggregates

Date: 2026-05-15

Status: Accepted

## Context

Pulse Check нужна серверная координация и агрегированная статистика, но privacy goal требует не хранить на сервере пользовательские персональные данные и персональные маппинги.

## Decision

Сервер хранит только technical identifiers, status identifiers/names там, где это нужно для общей статистики, и aggregate counters. Примеры: `entity_id`, `state_id`, `name`, `client_id`, `status_id` и `count`.

По умолчанию сервер не должен хранить attributes, notes, user fields или `ID -> name/person/note` mappings.

## Consequences

- Backend APIs остаются согласованными с privacy boundary.
- Product features, которым нужны readable names или notes, должны использовать local data или требовать отдельного explicit architecture decision.
- Aggregated manager/subordinate views можно строить вокруг statuses и counts без раскрытия конкретных entities.
- Developers должны считать logs, traces, metrics и support workflows частью той же data minimization boundary.
