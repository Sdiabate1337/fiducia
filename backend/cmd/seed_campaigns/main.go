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

	dbWrapper, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbWrapper.Close()
	db := dbWrapper.Pool
	ctx := context.Background()

	// get all cabinets
	rows, err := db.Query(ctx, "SELECT id, name FROM cabinets")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var cabinets []struct {
		ID   uuid.UUID
		Name string
	}

	for rows.Next() {
		var c struct {
			ID   uuid.UUID
			Name string
		}
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			log.Fatal(err)
		}
		cabinets = append(cabinets, c)
	}

	// For each cabinet, check if campaign exists
	for _, cab := range cabinets {
		var count int
		err := db.QueryRow(ctx, "SELECT count(*) FROM campaigns WHERE cabinet_id = $1", cab.ID).Scan(&count)
		if err != nil {
			log.Printf("Error checking cabinet %s: %v", cab.Name, err)
			continue
		}

		if count > 0 {
			log.Printf("Cabinet %s already has %d campaigns. Skipping.", cab.Name, count)
			continue
		}

		// Create Campaign
		log.Printf("Seeding campaign for cabinet: %s", cab.Name)
		campID := uuid.New()
		_, err = db.Exec(ctx, `
			INSERT INTO campaigns (id, cabinet_id, name, trigger_type, is_active, quiet_hours_enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, campID, cab.ID, "Smart Sequence v2", "on_pending", true, true, time.Now(), time.Now())
		if err != nil {
			log.Printf("Failed to insert campaign: %v", err)
			continue
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
			{2, 48, "voice", `{"script": "Bonjour, c'est votre comptable..."}`},
		}

		for _, s := range steps {
			_, err := db.Exec(ctx, `
				INSERT INTO campaign_steps (id, campaign_id, step_order, delay_hours, channel, template_id, config, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, uuid.New(), campID, s.Order, s.Delay, s.Channel, "default", s.Config, time.Now())
			if err != nil {
				log.Printf("Failed to insert step: %v", err)
			}
		}
	}
	log.Println("Done.")
}
