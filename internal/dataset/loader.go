package dataset

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

// Ref is a single reference vector with its label.
type Ref struct {
	Vector [14]float32
	Fraud  bool
}

type rawRef struct {
	Vector [14]float32 `json:"vector"`
	Label  string      `json:"label"`
}

// LoadRefs loads and decompresses references.json.gz into memory.
func LoadRefs(path string) ([]Ref, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var raw []rawRef
	if err := json.NewDecoder(gz).Decode(&raw); err != nil {
		return nil, err
	}

	refs := make([]Ref, len(raw))
	for i, r := range raw {
		refs[i] = Ref{Vector: r.Vector, Fraud: r.Label == "fraud"}
	}
	return refs, nil
}

// LoadMCCRisk loads mcc_risk.json.
func LoadMCCRisk(path string) (map[string]float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m map[string]float32
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

// Normalization holds the constants from normalization.json.
type Normalization struct {
	MaxAmount            float32 `json:"max_amount"`
	MaxInstallments      float32 `json:"max_installments"`
	AmountVsAvgRatio     float32 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float32 `json:"max_minutes"`
	MaxKm                float32 `json:"max_km"`
	MaxTxCount24h        float32 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float32 `json:"max_merchant_avg_amount"`
}

// LoadNormalization loads normalization.json.
func LoadNormalization(path string) (Normalization, error) {
	f, err := os.Open(path)
	if err != nil {
		return Normalization{}, err
	}
	defer f.Close()

	var n Normalization
	if err := json.NewDecoder(f).Decode(&n); err != nil {
		return Normalization{}, err
	}
	return n, nil
}
