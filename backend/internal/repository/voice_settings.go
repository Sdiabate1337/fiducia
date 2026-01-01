package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VoiceSetting struct {
	ID             uuid.UUID `json:"id"`
	CollaboratorID uuid.UUID `json:"collaborator_id"`
	VoiceID        string    `json:"voice_id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type VoiceSettingsRepository struct {
	db *pgxpool.Pool
}

func NewVoiceSettingsRepository(db *pgxpool.Pool) *VoiceSettingsRepository {
	return &VoiceSettingsRepository{db: db}
}

func (r *VoiceSettingsRepository) Create(ctx context.Context, setting *VoiceSetting) error {
	query := `
		INSERT INTO voice_settings (collaborator_id, voice_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query, setting.CollaboratorID, setting.VoiceID, setting.Name).
		Scan(&setting.ID, &setting.CreatedAt, &setting.UpdatedAt)
}

func (r *VoiceSettingsRepository) GetByCollaboratorID(ctx context.Context, collaboratorID uuid.UUID) (*VoiceSetting, error) {
	query := `
		SELECT id, collaborator_id, voice_id, name, created_at, updated_at
		FROM voice_settings
		WHERE collaborator_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var setting VoiceSetting
	err := r.db.QueryRow(ctx, query, collaboratorID).Scan(
		&setting.ID,
		&setting.CollaboratorID,
		&setting.VoiceID,
		&setting.Name,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}
