# ADR-0001: Local-first and Bring Your Own Backup

Date: 2026-05-15

Status: Accepted

## Context

Pulse Check может использоваться с пользовательскими названиями, заметками, полями, атрибутами и потенциально сведениями о других людях. Продукт должен минимизировать доступ сервера к этим данным и сохранять понятную границу ответственности: пользователь сам решает, что хранить локально и когда делать export/import.

## Decision

Pulse Check строится как local-first приложение. User-owned data по умолчанию остаются на устройстве пользователя. Backup следует модели Bring Your Own Backup: пользователь явно экспортирует зашифрованный backup, хранит его самостоятельно и импортирует при необходимости.

Планируемый backup archive содержит `manifest.json`, `data.json`, опциональный `attachments/` и `checksums.json`. Для ZIP планируется `fflate`, для валидации - Zod, для checksums - Web Crypto API/SHA-256.

## Consequences

- Default server boundary становится меньше и менее чувствительным.
- Приложение должно инвестировать в local persistence, export/import UX, backup integrity и понятные recovery instructions.
- Пользователь может потерять локальные данные, если не создает и не хранит backup.
- Browser storage persistence помогает, но не заменяет user-owned backup.
