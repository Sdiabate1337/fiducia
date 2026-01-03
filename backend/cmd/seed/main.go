package main

import (
	"context"
	"log"
	"time"

	"github.com/fiducia/backend/internal/config"
	"github.com/fiducia/backend/internal/database"
	"github.com/google/uuid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use internal database package to get migration support
	dbWrapper, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbWrapper.Close()

	// Run migrations to ensure schema exists (fixes "relation does not exist" error)
	if err := dbWrapper.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	db := dbWrapper.Pool

	ctx := context.Background()
	cabinetID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	campaignID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// Check if campaign exists
	var exists bool
	err = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM campaigns WHERE id=$1)", campaignID).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check campaign: %v", err)
	}

	if exists {
		log.Println("Campaign already exists. Skipping.")
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Failed to begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	// Create Campaign
	log.Println("Creating Smart Sequence Campaign...")
	_, err = tx.Exec(ctx, `
		INSERT INTO campaigns (id, cabinet_id, name, trigger_type, is_active, quiet_hours_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, campaignID, cabinetID, "Smart Sequence v2", "on_pending", true, true, time.Now(), time.Now())
	if err != nil {
		log.Fatalf("Failed to insert campaign: %v", err)
	}

	// Steps
	steps := []struct {
		Order   int
		Delay   int
		Channel string
		Config  string
	}{
		{1, 0, "email", `{"subject": "Justificatif manquant", "body": "Bonjour, merci de nous envoyer le justificatif."}`},
		{1, 0, "whatsapp", `{"template": "request_doc", "params": {}}`},
		{2, 48, "voice", `{"script": "Bonjour, c'est votre comptable..."}`}, // J+2 = 48h
		{3, 72, "notification", `{"message": "Escalade interne"}`},          // J+3 (relative to start? No, steps are usually sequential delay)
	}

	for _, s := range steps {
		_, err := tx.Exec(ctx, `
			INSERT INTO campaign_steps (id, campaign_id, step_order, delay_hours, channel, template_id, config, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, uuid.New(), campaignID, s.Order, s.Delay, s.Channel, "default", s.Config, time.Now())
		if err != nil {
			log.Fatalf("Failed to insert step: %v", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	log.Println("Campaign seeded successfully!")
}
