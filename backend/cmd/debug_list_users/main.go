package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/fiducia/backend/internal/config"
	"github.com/fiducia/backend/internal/database"
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

	rows, err := dbWrapper.Pool.Query(context.Background(), "SELECT id, email, full_name, cabinet_id, role FROM users")
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tEMAIL\tNAME\tCABINET_ID\tROLE")
	fmt.Fprintln(w, "--\t-----\t----\t----------\t----")

	for rows.Next() {
		var id, email, fullName, cabinetID, role string
		// Handle potential NULLs if any (scanning into strings usually works if not null, but let's assume they exist)
		if err := rows.Scan(&id, &email, &fullName, &cabinetID, &role); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, email, fullName, cabinetID, role)
	}
	w.Flush()
}
