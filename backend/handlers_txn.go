package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ── spending / transaction-feed endpoints ──────────────────────────────
// Data source in this build is a dummy partner-bank feed (see txn_seed.go).
// The campus never queries a bank account directly.

func (a *App) txSeed(w http.ResponseWriter, r *http.Request) {
	var students []Student
	var accounts []Account
	var disbs []Disbursement
	if err := a.db.Find(&students).Error; err != nil || len(students) == 0 {
		writeErr(w, http.StatusBadRequest, "seed penerima dulu (POST /api/seed/load)")
		return
	}
	a.db.Find(&accounts)
	a.db.Find(&disbs)

	data := generateTxns(students, accounts, disbs)
	if err := a.db.Exec("DELETE FROM transactions").Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "wipe transactions")
		return
	}
	if err := a.db.Exec("DELETE FROM spend_findings").Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "wipe spend findings")
		return
	}
	if err := a.db.CreateInBatches(data.Transactions, 200).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "insert transactions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"transactions": len(data.Transactions),
		"recipients":   len(students),
		"planted":      len(data.Labels),
		"source":       "dummy partner-bank feed",
	})
}

func (a *App) txDetect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var students []Student
	var txns []Transaction
	a.db.Find(&students)
	a.db.Find(&txns)
	if len(txns) == 0 {
		writeErr(w, http.StatusBadRequest, "belum ada transaksi — POST /api/tx/seed dulu")
		return
	}
	findings := detectSpending(students, txns)

	if err := a.db.Exec("DELETE FROM spend_findings").Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "clean findings")
		return
	}
	byRule := map[string]int{}
	for i := range findings {
		for _, rule := range strings.Split(findings[i].Rules, ",") {
			byRule[rule]++
		}
	}
	if len(findings) > 0 {
		if err := a.db.CreateInBatches(findings, 100).Error; err != nil {
			writeErr(w, http.StatusInternalServerError, "insert findings")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"findings":   len(findings),
		"by_rule":    byRule,
		"runtime_ms": time.Since(start).Milliseconds(),
	})
}

func (a *App) txFindings(w http.ResponseWriter, r *http.Request) {
	var rows []SpendFinding
	q := a.db.Order("risk_score desc")
	if b := r.URL.Query().Get("band"); b != "" {
		q = q.Where("band = ?", b)
	}
	if err := q.Find(&rows).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "query findings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"findings": rows})
}

func (a *App) txFindingDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var f SpendFinding
	if err := a.db.First(&f, "id = ?", id).Error; err != nil {
		writeErr(w, http.StatusNotFound, "finding tidak ditemukan")
		return
	}
	var txns []Transaction
	a.db.Where("student_id = ?", f.StudentID).Order("tanggal asc").Limit(60).Find(&txns)
	writeJSON(w, http.StatusOK, map[string]any{"finding": f, "transactions": txns})
}

func (a *App) txStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		ToStatus string `json:"to_status"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	allowed := map[string]map[string]bool{
		"baru":          {"diverifikasi": true, "salah_deteksi": true},
		"diverifikasi":  {"selesai": true, "salah_deteksi": true},
		"salah_deteksi": {"selesai": true},
	}
	var f SpendFinding
	if err := a.db.First(&f, "id = ?", id).Error; err != nil {
		writeErr(w, http.StatusNotFound, "finding tidak ditemukan")
		return
	}
	if !allowed[f.Status][in.ToStatus] {
		writeErr(w, http.StatusConflict, "transisi tidak valid: "+f.Status+" → "+in.ToStatus)
		return
	}
	err := a.db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&SpendFinding{}).Where("id = ?", f.ID).Update("status", in.ToStatus).Error
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "update status")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": in.ToStatus})
}

func (a *App) txStats(w http.ResponseWriter, r *http.Request) {
	var txCount, spendFindings, high, salah int64
	a.db.Raw("SELECT count(*) FROM transactions").Scan(&txCount)
	a.db.Raw("SELECT count(*) FROM spend_findings").Scan(&spendFindings)
	a.db.Raw("SELECT count(*) FROM spend_findings WHERE band = ?", BandPrioritas).Scan(&high)
	a.db.Raw("SELECT count(*) FROM spend_findings WHERE status = ?", "salah_deteksi").Scan(&salah)
	writeJSON(w, http.StatusOK, map[string]any{
		"transactions":     txCount,
		"findings":         spendFindings,
		"prioritas_tinggi": high,
		"salah_deteksi":    salah,
	})
}

func (a *App) txEval(w http.ResponseWriter, r *http.Request) {
	// Rebuild labels deterministically and measure coverage of the planted set.
	var students []Student
	var accounts []Account
	var disbs []Disbursement
	a.db.Find(&students)
	a.db.Find(&accounts)
	a.db.Find(&disbs)
	labels := generateTxns(students, accounts, disbs).Labels

	var rows []SpendFinding
	a.db.Order("risk_score desc").Find(&rows)
	got := map[string]string{}
	for _, f := range rows {
		got[f.NIM] = f.Rules
	}
	covered, hitsTop := 0, 0
	for _, l := range labels {
		rules, ok := got[l.NIM]
		if !ok {
			continue
		}
		all := true
		for _, want := range l.Rules {
			if !strings.Contains(rules, want) {
				all = false
				break
			}
		}
		if all {
			covered++
		}
	}
	top := rows
	if len(top) > 10 {
		top = top[:10]
	}
	planted := map[string]bool{}
	for _, l := range labels {
		planted[l.NIM] = true
	}
	for _, f := range top {
		if planted[f.NIM] {
			hitsTop++
		}
	}
	precision := 0.0
	if len(top) > 0 {
		precision = float64(hitsTop) / float64(len(top))
	}
	recall := 0.0
	if len(labels) > 0 {
		recall = float64(covered) / float64(len(labels))
	}
	fp, clean := 0, 0
	for _, f := range rows {
		if planted[f.NIM] {
			continue
		}
		fp++
	}
	clean = len(students) - len(labels)
	fpRate := 0.0
	if clean > 0 {
		fpRate = float64(fp) / float64(clean)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"precision_at_10":     precision,
		"recall":              recall,
		"covered":             covered,
		"planted":             len(labels),
		"false_positives":     fp,
		"false_positive_rate": fpRate,
		"methodology":         "transaksi sintetis (dummy feed bank), seed 42",
	})
}

// ── bank statement view (dummy feed) ───────────────────────────────────
// Returns the full mutation for one recipient with a running balance and the
// per-day / per-category series the UI draws as charts.

type statementDay struct {
	Tanggal string `json:"tanggal"`
	Keluar  int64  `json:"keluar"`
	Masuk   int64  `json:"masuk"`
}

type statementRow struct {
	ID       uint   `json:"id"`
	Tanggal  string `json:"tanggal"`
	Merchant string `json:"merchant"`
	Kategori string `json:"kategori"`
	Nominal  int64  `json:"nominal"`
	Arah     string `json:"arah"`
	Channel  string `json:"channel"`
	Saldo    int64  `json:"saldo"`
}

func (a *App) txStatement(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var s Student
	if err := a.db.First(&s, "id = ?", id).Error; err != nil {
		writeErr(w, http.StatusNotFound, "mahasiswa tidak ditemukan")
		return
	}
	var d Disbursement
	a.db.Where("student_id = ?", s.ID).First(&d)

	var txns []Transaction
	if err := a.db.Where("student_id = ?", s.ID).Order("tanggal asc, id asc").Find(&txns).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "query transaksi")
		return
	}
	if len(txns) == 0 {
		writeErr(w, http.StatusNotFound, "belum ada mutasi — jalankan POST /api/tx/seed")
		return
	}

	start := txns[0].Tanggal.AddDate(0, 0, -1)
	rows := make([]statementRow, 0, len(txns)+1)
	balance := d.Nominal
	rows = append(rows, statementRow{
		ID: 0, Tanggal: start.Format("2006-01-02"), Merchant: "PENCAIRAN KIP-K " + d.Periode,
		Kategori: "masuk", Nominal: d.Nominal, Arah: "masuk", Channel: "transfer", Saldo: balance,
	})

	byDay := map[string]*statementDay{}
	byCat := map[string]int64{}
	var totalKeluar int64
	for _, t := range txns {
		if t.Arah == "masuk" {
			balance += t.Nominal
			key := t.Tanggal.Format("2006-01-02")
			if byDay[key] == nil {
				byDay[key] = &statementDay{Tanggal: key}
			}
			byDay[key].Masuk += t.Nominal
			rows = append(rows, statementRow{
				ID: t.ID, Tanggal: key, Merchant: t.Merchant, Kategori: t.Kategori,
				Nominal: t.Nominal, Arah: t.Arah, Channel: t.Channel, Saldo: balance,
			})
			continue
		}
		balance -= t.Nominal
		totalKeluar += t.Nominal
		byCat[t.Kategori] += t.Nominal
		key := t.Tanggal.Format("2006-01-02")
		if byDay[key] == nil {
			byDay[key] = &statementDay{Tanggal: key}
		}
		byDay[key].Keluar += t.Nominal
		rows = append(rows, statementRow{
			ID: t.ID, Tanggal: key, Merchant: t.Merchant, Kategori: t.Kategori,
			Nominal: t.Nominal, Arah: t.Arah, Channel: t.Channel, Saldo: balance,
		})
	}

	days := make([]statementDay, 0, len(byDay))
	for _, v := range byDay {
		days = append(days, *v)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Tanggal < days[j].Tanggal })

	type catRow struct {
		Kategori string `json:"kategori"`
		Total    int64  `json:"total"`
	}
	cats := make([]catRow, 0, len(byCat))
	for k, v := range byCat {
		cats = append(cats, catRow{Kategori: k, Total: v})
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].Total > cats[j].Total })

	writeJSON(w, http.StatusOK, map[string]any{
		"student":       s,
		"disbursement":  d,
		"rows":          rows,
		"days":          days,
		"by_category":   cats,
		"total_keluar":  totalKeluar,
		"saldo_akhir":   balance,
		"source":        "dummy partner-bank feed",
	})
}
