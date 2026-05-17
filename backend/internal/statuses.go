package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	StatusRoleOwner       = "status_owner"
	StatusRoleParticipant = "participant"
	DefaultStatusLimit    = 50
	StatusNameMaxLength   = 40
)

var (
	ErrStatusForbidden = errors.New("status modification forbidden")
	ErrStatusLimit     = errors.New("status limit reached")
	ErrStatusNotFound  = errors.New("status not found")
	ErrInvalidStatus   = errors.New("invalid status")

	hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

type Status struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	BorderColor     string    `json:"border_color"`
	BackgroundColor string    `json:"background_color"`
	TextColor       string    `json:"text_color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StatusSet struct {
	ID          string   `json:"id"`
	OwnerUserID string   `json:"owner_user_id"`
	Role        string   `json:"role"`
	Statuses    []Status `json:"statuses"`
}

type StatusInput struct {
	Name            string `json:"name"`
	BorderColor     string `json:"border_color"`
	BackgroundColor string `json:"background_color"`
	TextColor       string `json:"text_color"`
}

type StatusStore interface {
	GetStatusSet(context.Context, string) (StatusSet, error)
	ListStatuses(context.Context, string) ([]Status, error)
	CreateStatus(context.Context, string, StatusInput) (Status, error)
	UpdateStatus(context.Context, string, string, StatusInput) (Status, error)
	DeleteStatus(context.Context, string, string) error
	Close() error
}

func validateStatusInput(input StatusInput) (StatusInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.BorderColor = strings.TrimSpace(input.BorderColor)
	input.BackgroundColor = strings.TrimSpace(input.BackgroundColor)
	input.TextColor = strings.TrimSpace(input.TextColor)

	if input.Name == "" {
		return input, fmt.Errorf("%w: name_required", ErrInvalidStatus)
	}
	if len([]rune(input.Name)) > StatusNameMaxLength {
		return input, fmt.Errorf("%w: name_too_long", ErrInvalidStatus)
	}
	if !hexColorPattern.MatchString(input.BorderColor) {
		return input, fmt.Errorf("%w: invalid_border_color", ErrInvalidStatus)
	}
	if !hexColorPattern.MatchString(input.BackgroundColor) {
		return input, fmt.Errorf("%w: invalid_background_color", ErrInvalidStatus)
	}
	if !hexColorPattern.MatchString(input.TextColor) {
		return input, fmt.Errorf("%w: invalid_text_color", ErrInvalidStatus)
	}

	return input, nil
}

type memoryStatusStore struct {
	mu          sync.Mutex
	maxStatuses int
	memberships map[string]statusSetMembership
	sets        map[string]statusSetRecord
	statuses    map[string][]Status
}

type statusSetMembership struct {
	UserID      string
	StatusSetID string
	OwnerUserID string
}

type statusSetRecord struct {
	ID          string
	OwnerUserID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewMemoryStatusStore(maxStatuses int) StatusStore {
	if maxStatuses <= 0 {
		maxStatuses = DefaultStatusLimit
	}

	return &memoryStatusStore{
		maxStatuses: maxStatuses,
		memberships: map[string]statusSetMembership{},
		sets:        map[string]statusSetRecord{},
		statuses:    map[string][]Status{},
	}
}

func (s *memoryStatusStore) GetStatusSet(_ context.Context, userID string) (StatusSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	membership := s.ensureMembershipLocked(userID)
	return s.statusSetLocked(membership), nil
}

func (s *memoryStatusStore) ListStatuses(ctx context.Context, userID string) ([]Status, error) {
	statusSet, err := s.GetStatusSet(ctx, userID)
	if err != nil {
		return nil, err
	}

	return statusSet.Statuses, nil
}

func (s *memoryStatusStore) CreateStatus(_ context.Context, userID string, input StatusInput) (Status, error) {
	input, err := validateStatusInput(input)
	if err != nil {
		return Status{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	membership := s.ensureMembershipLocked(userID)
	if membership.OwnerUserID != userID {
		return Status{}, ErrStatusForbidden
	}
	if len(s.statuses[membership.StatusSetID]) >= s.maxStatuses {
		return Status{}, ErrStatusLimit
	}

	now := time.Now().UTC()
	status := Status{
		ID:              uuid.NewString(),
		Name:            input.Name,
		BorderColor:     input.BorderColor,
		BackgroundColor: input.BackgroundColor,
		TextColor:       input.TextColor,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.statuses[membership.StatusSetID] = append(s.statuses[membership.StatusSetID], status)
	s.touchStatusSetLocked(membership.StatusSetID, now)

	return status, nil
}

func (s *memoryStatusStore) UpdateStatus(_ context.Context, userID string, statusID string, input StatusInput) (Status, error) {
	input, err := validateStatusInput(input)
	if err != nil {
		return Status{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	membership := s.ensureMembershipLocked(userID)
	if membership.OwnerUserID != userID {
		return Status{}, ErrStatusForbidden
	}

	now := time.Now().UTC()
	for index, status := range s.statuses[membership.StatusSetID] {
		if status.ID != statusID {
			continue
		}

		status.Name = input.Name
		status.BorderColor = input.BorderColor
		status.BackgroundColor = input.BackgroundColor
		status.TextColor = input.TextColor
		status.UpdatedAt = now
		s.statuses[membership.StatusSetID][index] = status
		s.touchStatusSetLocked(membership.StatusSetID, now)

		return status, nil
	}

	return Status{}, ErrStatusNotFound
}

func (s *memoryStatusStore) DeleteStatus(_ context.Context, userID string, statusID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	membership := s.ensureMembershipLocked(userID)
	if membership.OwnerUserID != userID {
		return ErrStatusForbidden
	}

	for index, status := range s.statuses[membership.StatusSetID] {
		if status.ID != statusID {
			continue
		}

		s.statuses[membership.StatusSetID] = append(s.statuses[membership.StatusSetID][:index], s.statuses[membership.StatusSetID][index+1:]...)
		s.touchStatusSetLocked(membership.StatusSetID, time.Now().UTC())
		return nil
	}

	return ErrStatusNotFound
}

func (s *memoryStatusStore) Close() error {
	return nil
}

func (s *memoryStatusStore) ensureMembershipLocked(userID string) statusSetMembership {
	if membership, ok := s.memberships[userID]; ok {
		return membership
	}

	now := time.Now().UTC()
	statusSetID := uuid.NewString()
	s.sets[statusSetID] = statusSetRecord{
		ID:          statusSetID,
		OwnerUserID: userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	membership := statusSetMembership{
		UserID:      userID,
		StatusSetID: statusSetID,
		OwnerUserID: userID,
	}
	s.memberships[userID] = membership

	return membership
}

func (s *memoryStatusStore) statusSetLocked(membership statusSetMembership) StatusSet {
	role := StatusRoleParticipant
	if membership.UserID == membership.OwnerUserID {
		role = StatusRoleOwner
	}

	statuses := append([]Status(nil), s.statuses[membership.StatusSetID]...)

	return StatusSet{
		ID:          membership.StatusSetID,
		OwnerUserID: membership.OwnerUserID,
		Role:        role,
		Statuses:    statuses,
	}
}

func (s *memoryStatusStore) touchStatusSetLocked(statusSetID string, updatedAt time.Time) {
	statusSet := s.sets[statusSetID]
	statusSet.UpdatedAt = updatedAt
	s.sets[statusSetID] = statusSet
}

type PostgresStatusStore struct {
	db          *sql.DB
	maxStatuses int
	logger      *slog.Logger
}

func NewPostgresStatusStore(ctx context.Context, databaseURL string, maxStatuses int, logger *slog.Logger) (*PostgresStatusStore, error) {
	if maxStatuses <= 0 {
		maxStatuses = DefaultStatusLimit
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	store := &PostgresStatusStore{
		db:          db,
		maxStatuses: maxStatuses,
		logger:      logger,
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.initSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *PostgresStatusStore) GetStatusSet(ctx context.Context, userID string) (StatusSet, error) {
	membership, err := s.ensureMembership(ctx, userID)
	if err != nil {
		return StatusSet{}, err
	}

	statuses, err := s.listStatusesBySet(ctx, membership.StatusSetID)
	if err != nil {
		return StatusSet{}, err
	}

	role := StatusRoleParticipant
	if membership.UserID == membership.OwnerUserID {
		role = StatusRoleOwner
	}

	return StatusSet{
		ID:          membership.StatusSetID,
		OwnerUserID: membership.OwnerUserID,
		Role:        role,
		Statuses:    statuses,
	}, nil
}

func (s *PostgresStatusStore) ListStatuses(ctx context.Context, userID string) ([]Status, error) {
	membership, err := s.ensureMembership(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.listStatusesBySet(ctx, membership.StatusSetID)
}

func (s *PostgresStatusStore) CreateStatus(ctx context.Context, userID string, input StatusInput) (Status, error) {
	input, err := validateStatusInput(input)
	if err != nil {
		return Status{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Status{}, err
	}
	defer rollbackTx(tx)

	membership, err := s.ensureMembershipTx(ctx, tx, userID)
	if err != nil {
		return Status{}, err
	}
	if membership.OwnerUserID != userID {
		return Status{}, ErrStatusForbidden
	}

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM statuses WHERE status_set_id = $1`, membership.StatusSetID).Scan(&count); err != nil {
		return Status{}, err
	}
	if count >= s.maxStatuses {
		return Status{}, ErrStatusLimit
	}

	statusID := uuid.NewString()
	var status Status
	err = tx.QueryRowContext(ctx, `
		INSERT INTO statuses (id, status_set_id, name, border_color, background_color, text_color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())
		RETURNING id::text, name, border_color, background_color, text_color, created_at, updated_at
	`, statusID, membership.StatusSetID, input.Name, input.BorderColor, input.BackgroundColor, input.TextColor).Scan(
		&status.ID,
		&status.Name,
		&status.BorderColor,
		&status.BackgroundColor,
		&status.TextColor,
		&status.CreatedAt,
		&status.UpdatedAt,
	)
	if err != nil {
		return Status{}, err
	}

	if err := touchStatusSet(ctx, tx, membership.StatusSetID); err != nil {
		return Status{}, err
	}
	if err := tx.Commit(); err != nil {
		return Status{}, err
	}

	return status, nil
}

func (s *PostgresStatusStore) UpdateStatus(ctx context.Context, userID string, statusID string, input StatusInput) (Status, error) {
	input, err := validateStatusInput(input)
	if err != nil {
		return Status{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Status{}, err
	}
	defer rollbackTx(tx)

	membership, err := s.ensureMembershipTx(ctx, tx, userID)
	if err != nil {
		return Status{}, err
	}
	if membership.OwnerUserID != userID {
		return Status{}, ErrStatusForbidden
	}

	var status Status
	err = tx.QueryRowContext(ctx, `
		UPDATE statuses
		SET name = $1,
			border_color = $2,
			background_color = $3,
			text_color = $4,
			updated_at = now()
		WHERE id = $5 AND status_set_id = $6
		RETURNING id::text, name, border_color, background_color, text_color, created_at, updated_at
	`, input.Name, input.BorderColor, input.BackgroundColor, input.TextColor, statusID, membership.StatusSetID).Scan(
		&status.ID,
		&status.Name,
		&status.BorderColor,
		&status.BackgroundColor,
		&status.TextColor,
		&status.CreatedAt,
		&status.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Status{}, ErrStatusNotFound
	}
	if err != nil {
		return Status{}, err
	}

	if err := touchStatusSet(ctx, tx, membership.StatusSetID); err != nil {
		return Status{}, err
	}
	if err := tx.Commit(); err != nil {
		return Status{}, err
	}

	return status, nil
}

func (s *PostgresStatusStore) DeleteStatus(ctx context.Context, userID string, statusID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

	membership, err := s.ensureMembershipTx(ctx, tx, userID)
	if err != nil {
		return err
	}
	if membership.OwnerUserID != userID {
		return ErrStatusForbidden
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM statuses
		WHERE id = $1 AND status_set_id = $2
	`, statusID, membership.StatusSetID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrStatusNotFound
	}

	if err := touchStatusSet(ctx, tx, membership.StatusSetID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	if s.logger != nil {
		s.logger.Warn(
			"status deleted; TODO clear or reassign this status on entities that use it",
			slog.String("status_id", statusID),
		)
	}

	return nil
}

func (s *PostgresStatusStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStatusStore) initSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS status_sets (
			id UUID PRIMARY KEY,
			owner_user_id TEXT NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS status_set_memberships (
			user_id TEXT PRIMARY KEY,
			status_set_id UUID NOT NULL REFERENCES status_sets(id) ON DELETE CASCADE,
			owner_user_id TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		);

		CREATE INDEX IF NOT EXISTS status_set_memberships_status_set_id_idx
			ON status_set_memberships(status_set_id);

		CREATE TABLE IF NOT EXISTS statuses (
			id UUID PRIMARY KEY,
			status_set_id UUID NOT NULL REFERENCES status_sets(id) ON DELETE CASCADE,
			name TEXT NOT NULL CHECK (char_length(name) <= 40),
			border_color CHAR(7) NOT NULL CHECK (border_color ~ '^#[0-9A-Fa-f]{6}$'),
			background_color CHAR(7) NOT NULL CHECK (background_color ~ '^#[0-9A-Fa-f]{6}$'),
			text_color CHAR(7) NOT NULL CHECK (text_color ~ '^#[0-9A-Fa-f]{6}$'),
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);

		CREATE INDEX IF NOT EXISTS statuses_status_set_id_idx
			ON statuses(status_set_id);

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1
				FROM pg_constraint
				WHERE conname = 'statuses_name_length_check'
					AND conrelid = 'statuses'::regclass
			) THEN
				ALTER TABLE statuses
					ADD CONSTRAINT statuses_name_length_check CHECK (char_length(name) <= 40) NOT VALID;
			END IF;
		END $$;
	`)

	return err
}

func (s *PostgresStatusStore) ensureMembership(ctx context.Context, userID string) (statusSetMembership, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return statusSetMembership{}, err
	}
	defer rollbackTx(tx)

	membership, err := s.ensureMembershipTx(ctx, tx, userID)
	if err != nil {
		return statusSetMembership{}, err
	}
	if err := tx.Commit(); err != nil {
		return statusSetMembership{}, err
	}

	return membership, nil
}

func (s *PostgresStatusStore) ensureMembershipTx(ctx context.Context, tx *sql.Tx, userID string) (statusSetMembership, error) {
	var membership statusSetMembership
	err := tx.QueryRowContext(ctx, `
		SELECT user_id, status_set_id::text, owner_user_id
		FROM status_set_memberships
		WHERE user_id = $1
	`, userID).Scan(&membership.UserID, &membership.StatusSetID, &membership.OwnerUserID)
	if err == nil {
		return membership, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return statusSetMembership{}, err
	}

	statusSetID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO status_sets (id, owner_user_id, created_at, updated_at)
		VALUES ($1, $2, now(), now())
		ON CONFLICT (owner_user_id) DO NOTHING
	`, statusSetID, userID)
	if err != nil {
		return statusSetMembership{}, err
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id::text
		FROM status_sets
		WHERE owner_user_id = $1
	`, userID).Scan(&statusSetID)
	if err != nil {
		return statusSetMembership{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO status_set_memberships (user_id, status_set_id, owner_user_id, created_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id) DO NOTHING
	`, userID, statusSetID, userID)
	if err != nil {
		return statusSetMembership{}, err
	}

	err = tx.QueryRowContext(ctx, `
		SELECT user_id, status_set_id::text, owner_user_id
		FROM status_set_memberships
		WHERE user_id = $1
	`, userID).Scan(&membership.UserID, &membership.StatusSetID, &membership.OwnerUserID)
	if err != nil {
		return statusSetMembership{}, err
	}

	return membership, nil
}

func (s *PostgresStatusStore) listStatusesBySet(ctx context.Context, statusSetID string) ([]Status, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id::text, name, border_color, background_color, text_color, created_at, updated_at
		FROM statuses
		WHERE status_set_id = $1
		ORDER BY created_at ASC, id ASC
	`, statusSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statuses := []Status{}
	for rows.Next() {
		var status Status
		if err := rows.Scan(
			&status.ID,
			&status.Name,
			&status.BorderColor,
			&status.BackgroundColor,
			&status.TextColor,
			&status.CreatedAt,
			&status.UpdatedAt,
		); err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return statuses, nil
}

func touchStatusSet(ctx context.Context, tx *sql.Tx, statusSetID string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE status_sets
		SET updated_at = now()
		WHERE id = $1
	`, statusSetID)

	return err
}

func rollbackTx(tx *sql.Tx) {
	_ = tx.Rollback()
}
