package fraud

import (
	"math"
	"time"

	"github.com/Menon04/rinha-de-backend-2026-golang/internal/dataset"
)

// Request mirrors the POST /fraud-score payload.
type Request struct {
	ID          string      `json:"id"`
	Transaction Transaction `json:"transaction"`
	Customer    Customer    `json:"customer"`
	Merchant    Merchant    `json:"merchant"`
	Terminal    Terminal    `json:"terminal"`
	LastTx      *LastTx     `json:"last_transaction"`
}

type Transaction struct {
	Amount      float32 `json:"amount"`
	Installments int    `json:"installments"`
	RequestedAt string  `json:"requested_at"`
}

type Customer struct {
	AvgAmount       float32  `json:"avg_amount"`
	TxCount24h      int      `json:"tx_count_24h"`
	KnownMerchants  []string `json:"known_merchants"`
}

type Merchant struct {
	ID        string  `json:"id"`
	MCC       string  `json:"mcc"`
	AvgAmount float32 `json:"avg_amount"`
}

type Terminal struct {
	IsOnline    bool    `json:"is_online"`
	CardPresent bool    `json:"card_present"`
	KmFromHome  float32 `json:"km_from_home"`
}

type LastTx struct {
	Timestamp     string  `json:"timestamp"`
	KmFromCurrent float32 `json:"km_from_current"`
}

func clamp(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Vectorize converts a fraud request into a 14-dimensional vector.
func Vectorize(req *Request, norm dataset.Normalization, mccRisk map[string]float32) [14]float32 {
	var v [14]float32

	// 0: amount
	v[0] = clamp(req.Transaction.Amount / norm.MaxAmount)

	// 1: installments
	v[1] = clamp(float32(req.Transaction.Installments) / norm.MaxInstallments)

	// 2: amount_vs_avg
	if req.Customer.AvgAmount > 0 {
		v[2] = clamp((req.Transaction.Amount / req.Customer.AvgAmount) / norm.AmountVsAvgRatio)
	}

	// 3: hour_of_day, 4: day_of_week
	if t, err := time.Parse(time.RFC3339, req.Transaction.RequestedAt); err == nil {
		v[3] = float32(t.UTC().Hour()) / 23.0
		dow := int(t.UTC().Weekday()) - 1 // Monday=0, Sunday=6
		if dow < 0 {
			dow = 6
		}
		v[4] = float32(dow) / 6.0
	}

	// 5: minutes_since_last_tx, 6: km_from_last_tx
	if req.LastTx == nil {
		v[5] = -1
		v[6] = -1
	} else {
		if t1, err1 := time.Parse(time.RFC3339, req.Transaction.RequestedAt); err1 == nil {
			if t2, err2 := time.Parse(time.RFC3339, req.LastTx.Timestamp); err2 == nil {
				mins := float32(math.Abs(t1.Sub(t2).Minutes()))
				v[5] = clamp(mins / norm.MaxMinutes)
			}
		}
		v[6] = clamp(req.LastTx.KmFromCurrent / norm.MaxKm)
	}

	// 7: km_from_home
	v[7] = clamp(req.Terminal.KmFromHome / norm.MaxKm)

	// 8: tx_count_24h
	v[8] = clamp(float32(req.Customer.TxCount24h) / norm.MaxTxCount24h)

	// 9: is_online
	if req.Terminal.IsOnline {
		v[9] = 1
	}

	// 10: card_present
	if req.Terminal.CardPresent {
		v[10] = 1
	}

	// 11: unknown_merchant
	v[11] = 1
	for _, m := range req.Customer.KnownMerchants {
		if m == req.Merchant.ID {
			v[11] = 0
			break
		}
	}

	// 12: mcc_risk
	if risk, ok := mccRisk[req.Merchant.MCC]; ok {
		v[12] = risk
	} else {
		v[12] = 0.5
	}

	// 13: merchant_avg_amount
	v[13] = clamp(req.Merchant.AvgAmount / norm.MaxMerchantAvgAmount)

	return v
}
