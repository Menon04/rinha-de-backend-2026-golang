package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Menon04/rinha-de-backend-2026-golang/internal/dataset"
	"github.com/Menon04/rinha-de-backend-2026-golang/internal/fraud"
)

type App struct {
	refs    []dataset.Ref
	norm    dataset.Normalization
	mccRisk map[string]float32
}

func New(refs []dataset.Ref, norm dataset.Normalization, mccRisk map[string]float32) *App {
	return &App{refs: refs, norm: norm, mccRisk: mccRisk}
}

func (a *App) Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type fraudResponse struct {
	Approved   bool    `json:"approved"`
	FraudScore float32 `json:"fraud_score"`
}

func (a *App) FraudScore(w http.ResponseWriter, r *http.Request) {
	var req fraud.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	vec := fraud.Vectorize(&req, a.norm, a.mccRisk)
	score, approved := fraud.Score(vec, a.refs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fraudResponse{Approved: approved, FraudScore: score})
}
