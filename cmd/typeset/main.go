package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-typeset/internal/server"
	"github.com/stockyard-dev/stockyard-typeset/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9240"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./typeset-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("typeset: open database: %v", err)
	}
	defer db.Close()

	srv := server.New(db)

	fmt.Printf("\n  Typeset — Self-hosted documentation generator\n")
	fmt.Printf("  ─────────────────────────────────\n")
	fmt.Printf("  Dashboard:  http://localhost:%s/ui\n", port)
	fmt.Printf("  API:        http://localhost:%s/api\n", port)
	fmt.Printf("  Data:       %s\n", dataDir)
	fmt.Printf("  ─────────────────────────────────\n\n")

	log.Printf("typeset: listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("typeset: %v", err)
	}
}
