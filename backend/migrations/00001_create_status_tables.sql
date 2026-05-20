-- +goose Up
CREATE TABLE status_sets (
	id UUID PRIMARY KEY,
	owner_user_id UUID NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE status_set_memberships (
	user_id UUID PRIMARY KEY,
	status_set_id UUID NOT NULL REFERENCES status_sets(id) ON DELETE CASCADE,
	owner_user_id UUID NOT NULL,
	created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX status_set_memberships_status_set_id_idx
	ON status_set_memberships(status_set_id);

CREATE TABLE statuses (
	id UUID PRIMARY KEY,
	status_set_id UUID NOT NULL REFERENCES status_sets(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	border_color CHAR(7) NOT NULL,
	background_color CHAR(7) NOT NULL,
	text_color CHAR(7) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	CONSTRAINT statuses_name_length_check CHECK (char_length(name) <= 40),
	CONSTRAINT statuses_border_color_check CHECK (border_color ~ '^#[0-9A-Fa-f]{6}$'),
	CONSTRAINT statuses_background_color_check CHECK (background_color ~ '^#[0-9A-Fa-f]{6}$'),
	CONSTRAINT statuses_text_color_check CHECK (text_color ~ '^#[0-9A-Fa-f]{6}$')
);

CREATE INDEX statuses_status_set_id_idx
	ON statuses(status_set_id);

-- +goose Down
DROP TABLE statuses;
DROP TABLE status_set_memberships;
DROP TABLE status_sets;
