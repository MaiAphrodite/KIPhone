package main

import "time"

// Transaction data arrives from the partner-bank feed (dummy in this build).
// The campus never queries a student's account directly; it receives this
// stream under a data-sharing agreement.

type Transaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StudentID uint      `gorm:"index" json:"student_id"`
	Tanggal   time.Time `json:"tanggal"`
	Merchant  string    `gorm:"size:120" json:"merchant"`
	Kategori  string    `gorm:"size:24;index" json:"kategori"` // makan|transport|pulsa|belanja|mewah|taruhan|tunai|transfer|lainnya
	Nominal   int64     `json:"nominal"`
	Arah      string    `gorm:"size:8" json:"arah"` // keluar | masuk
	Channel   string    `gorm:"size:16" json:"channel"` // qris | transfer | atm | debit | ewallet
}

type SpendFinding struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	StudentID   uint      `gorm:"index" json:"student_id"`
	NIM         string    `gorm:"size:20" json:"nim"`
	Nama        string    `gorm:"size:120" json:"nama"`
	Rules       string    `gorm:"size:20" json:"rules"`
	RiskScore   float64   `json:"risk_score"`
	Band        string    `gorm:"size:24" json:"band"`
	Explanation string    `gorm:"type:text" json:"explanation"`
	Evidence    string    `gorm:"type:jsonb" json:"evidence"`
	Status      string    `gorm:"size:24" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

var spendRuleWeights = map[string]float64{"T1": 0.50, "T2": 0.30, "T3": 0.15, "T4": 0.20}
