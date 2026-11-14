package main

import "time"

// Domain per PRD §5. One Finding per disbursement: Rule holds a comma list
// ("R2,R3") when several rules hit the same record.

type Student struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	NIM            string    `gorm:"index;size:20" json:"nim"`
	NIK            string    `gorm:"index;size:20" json:"nik"`
	Nama           string    `gorm:"size:120" json:"nama"`
	Prodi          string    `gorm:"size:80" json:"prodi"`
	StatusAkademik string    `gorm:"size:20" json:"status_akademik"` // aktif | cuti | nonaktif | lulus
	LastActiveAt   time.Time `json:"last_active_at"`
}

type Account struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StudentID     uint      `gorm:"index" json:"student_id"`
	NomorRekening string    `gorm:"index;size:40" json:"nomor_rekening"`
	NamaPemilik   string    `gorm:"size:120" json:"nama_pemilik"`
	CreatedAt     time.Time `json:"created_at"`
}

type Disbursement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	StudentID   uint      `gorm:"index" json:"student_id"`
	AccountID   uint      `json:"account_id"`
	SKNomor     string    `gorm:"size:60" json:"sk_nomor"`
	Periode     string    `gorm:"size:10" json:"periode"`
	Nominal     int64     `json:"nominal"`
	TanggalCair time.Time `json:"tanggal_cair"`
}

type Finding struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	DisbursementID uint      `gorm:"index" json:"disbursement_id"`
	Rule           string    `gorm:"size:12" json:"rule"` // "R1" | "R2,R3" | ...
	RiskScore      float64   `json:"risk_score"`
	Band           string    `gorm:"size:20" json:"band"` // rendah | perlu_verifikasi | prioritas_tinggi
	Explanation    string    `gorm:"type:text" json:"explanation"`
	Evidence       string    `gorm:"type:jsonb" json:"evidence"` // JSON object keyed by rule
	Status         string    `gorm:"size:20" json:"status"`      // baru | diverifikasi | salah_deteksi | selesai
	CreatedAt      time.Time `json:"created_at"`
}

type FindingEvent struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FindingID  uint      `gorm:"index" json:"finding_id"`
	FromStatus string    `gorm:"size:20" json:"from_status"`
	ToStatus   string    `gorm:"size:20" json:"to_status"`
	Actor      string    `gorm:"size:80" json:"actor"`
	Note       string    `gorm:"type:text" json:"note"`
	At         time.Time `json:"at"`
}
