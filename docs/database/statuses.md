# Таблица `statuses`:

Хранит определения статусов внутри конкретного набора: название, цвета и технические timestamps.

| Поле               | Тип           | Constraint | Ограничения                                                                                                 |
|--------------------|---------------|------------|-------------------------------------------------------------------------------------------------------------|
| `id`               | `UUID`        | not null   | Первичный ключ, генерируется backend-ом и не меняется                                                       |
| `status_set_id`    | `UUID`        | not null   | Внешний ключ на `status_sets.id` с `ON DELETE CASCADE`; индекс `statuses_status_set_id_idx`                 |
| `name`             | `TEXT`        | not null   | Не длиннее 40 символов; check constraint `char_length(name) <= 40`; constraint `statuses_name_length_check` |
| `border_color`     | `CHAR(7)`     | not null   | HEX-цвет в формате `#RRGGBB`; check constraint по regex `^#[0-9A-Fa-f]{6}$`                                 |
| `background_color` | `CHAR(7)`     | not null   | HEX-цвет в формате `#RRGGBB`; check constraint по regex `^#[0-9A-Fa-f]{6}$`                                 |
| `text_color`       | `CHAR(7)`     | not null   | HEX-цвет в формате `#RRGGBB`; check constraint по regex `^#[0-9A-Fa-f]{6}$`                                 |
| `created_at`       | `TIMESTAMPTZ` | not null   | Дата создания записи                                                                                        |
| `updated_at`       | `TIMESTAMPTZ` | not null   | Дата последнего изменения записи                                                                            |

Дополнительные заметки:

- `name` - ограничение 40 символов проверяется на frontend для мгновенной обратной связи и на backend как source of truth.
