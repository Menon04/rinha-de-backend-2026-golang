package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Menon04/rinha-de-backend-2026-golang/internal/dataset"
	"github.com/Menon04/rinha-de-backend-2026-golang/internal/handler"
)

func main() {
	refsPath := envOr("REFS_PATH", "/data/references.json.gz")
	mccPath := envOr("MCC_PATH", "/data/mcc_risk.json")
	normPath := envOr("NORM_PATH", "/data/normalization.json")

	log.Println("loading references...")
	refs, err := dataset.LoadRefs(refsPath)
	if err != nil {
		log.Fatalf("failed to load refs: %v", err)
	}
	log.Printf("loaded %d references", len(refs))

	mccRisk, err := dataset.LoadMCCRisk(mccPath)
	if err != nil {
		log.Fatalf("failed to load mcc_risk: %v", err)
	}

	norm, err := dataset.LoadNormalization(normPath)
	if err != nil {
		log.Fatalf("failed to load normalization: %v", err)
	}

	app := handler.New(refs, norm, mccRisk)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ready", app.Ready)
	mux.HandleFunc("POST /fraud-score", app.FraudScore)

	addr := envOr("ADDR", ":8080")
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
