package main

import (
	"context"
	"fmt"
	"log"
	
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

	// Find the line
	var id uuid.UUID
	var clientID *uuid.UUID
	var bankLabel string
	
	err = db.QueryRow(ctx, `
		SELECT id, client_id, bank_label 
		FROM pending_lines 
		WHERE bank_label LIKE '%LECLERC%' 
		LIMIT 1
	`).Scan(&id, &clientID, &bankLabel)
	
	if err != nil {
		log.Fatalf("Line not found: %v", err)
	}
	
	fmt.Printf("Found Line: %s (ID: %s)\n", bankLabel, id)
	if clientID != nil {
		fmt.Printf("Linked Client ID: %s\n", *clientID)
	} else {
		fmt.Printf("Linked Client ID: <nil>\n")
	}

	// Now try the GetByID query logic
	query := `
		SELECT 
			pl.id, c.id as client_id, c.name as client_name
		FROM pending_lines pl
		LEFT JOIN clients c ON pl.client_id = c.id
		WHERE pl.id = $1
	`
	var plID, cID *string, cName *string
	err = db.QueryRow(ctx, query, id).Scan(&plID, &cID, &cName)
	if err != nil {
		log.Fatalf("GetByID Query Failed: %v", err)
	}
	
	fmt.Printf("Join Result:\nLineID: %v\nClientID: %v\nClientName: %v\n", plID, cID, cName)
}
