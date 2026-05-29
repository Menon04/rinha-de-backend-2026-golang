package dataset

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

// Ref is a quantized reference vector with its label.
// Vectors are stored as uint8 (mapped from [0,1] → [0,255])
// to reduce memory ~4x vs float32.
type Ref struct {
	Vector [14]uint8
	Fraud  bool
}

func Quantize(v [14]float32) [14]uint8 {
	var u [14]uint8
	for i := 0; i < 14; i++ {
		s := v[i] * 255
		if s < 0 {
			s = 0
		}
		if s > 255 {
			s = 255
		}
		u[i] = uint8(s)
	}
	return u
}

type rawRef struct {
	Vector [14]float32 `json:"vector"`
	Label  string      `json:"label"`
}

// LoadRefs loads and decompresses references.json.gz into memory.
// LoadRefs loads and decompresses references.json.gz into memory using streaming
// to avoid holding two copies of the data simultaneously.
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

	dec := json.NewDecoder(gz)

	// read opening '['
	if _, err := dec.Token(); err != nil {
		return nil, err
	}

	refs := make([]Ref, 0, 3_000_000)
	var raw rawRef
	for dec.More() {
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		refs = append(refs, Ref{Vector: Quantize(raw.Vector), Fraud: raw.Label == "fraud"})
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
