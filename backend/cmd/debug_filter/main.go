package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fiducia/backend/internal/models"
	"github.com/fiducia/backend/internal/repository"
)

func main() {
	// Attempt to read DB URL from environment or use default from .env.example
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/fiducia?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.NewPendingLineRepository(pool)

	cabinetID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	pendingStatus := models.StatusPending

	filter := repository.PendingLineFilter{
		CabinetID: cabinetID,
		Status:    &pendingStatus,
		Limit:     10,
	}

	fmt.Printf("Testing List with Filter: %+v\n", filter)
	fmt.Printf("Status Pointer: %v, Value: %s\n", filter.Status, *filter.Status)

	list, err := repo.List(ctx, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "List failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d items\n", len(list.Items))
	for _, item := range list.Items {
		fmt.Printf("ID: %s, Status: %s\n", item.ID, item.Status)
		if item.Status != models.StatusPending {
			fmt.Printf("!!! MISMATCH DETECTED !!! Expected pending, got %s\n", item.Status)
		}
	}
}
