package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PRD §7: seed → detect → findings → lifecycle → stats → eval.

func (a *App) loadSeed(w http.ResponseWriter, r *http.Request) {
	data := generateSeed()
	if err := a.db.Exec("DELETE FROM finding_events").Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "wipe finding_events")
		return
	}
	for _, t := range []string{"findings", "disbursements", "accounts", "students"} {
		if err := a.db.Exec("DELETE FROM " + t).Error; err != nil {
			writeErr(w, http.StatusInternalServerError, "wipe "+t)
			return
		}
	}
	if err := a.db.CreateInBatches(data.Students, 100).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "insert students")
		return
	}
	if err := a.db.CreateInBatches(data.Accounts, 100).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "insert accounts")
		return
	}
	if err := a.db.CreateInBatches(data.Disbursements, 100).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "insert disbursements")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"students":      len(data.Students),
		"accounts":      len(data.Accounts),
		"disbursements": len(data.Disbursements),
		"planted":       len(data.Labels),
	})
}

func (a *App) runDetect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var students []Student
	var accounts []Account
	var disbs []Disbursement
	if err := a.db.Find(&students).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "load students")
		return
	}
	if err := a.db.Find(&accounts).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "load accounts")
		return
	}
	if err := a.db.Find(&disbs).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "load disbursements")
		return
	}
	if len(disbs) == 0 {
		writeErr(w, http.StatusBadRequest, "batch kosong — panggil POST /api/seed/load dulu")
		return
	}

	findings := detect(students, accounts, disbs)

	// Detection is a recompute: every run replaces the full finding set, so a
	// repeated run cannot duplicate rows or inflate stats.
	if err := a.db.Exec("DELETE FROM finding_events").Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "clean events")
		return
	}
	if err := a.db.Exec("DELETE FROM findings").Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "clean findings")
		return
	}
	byRule := map[string]int{}
	for i := range findings {
		for _, rule := range strings.Split(findings[i].Rule, ",") {
			byRule[rule]++
		}
	}
	if len(findings) > 0 {
		if err := a.db.CreateInBatches(findings, 100).Error; err != nil {
			writeErr(w, http.StatusInternalServerError, "insert findings")
			return
		}
		events := make([]FindingEvent, 0, len(findings))
		for _, f := range findings {
			events = append(events, FindingEvent{FindingID: f.ID, FromStatus: "", ToStatus: StatusBaru, Actor: "system", Note: "terdeteksi otomatis", At: time.Now()})
		}
		if err := a.db.CreateInBatches(events, 100).Error; err != nil {
			writeErr(w, http.StatusInternalServerError, "insert events")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"findings_created": len(findings),
		"by_rule":          byRule,
		"runtime_ms":       time.Since(start).Milliseconds(),
	})
}

type findingRow struct {
	Finding
	NIM          string `json:"nim"`
	Nama         string `json:"nama"`
	Prodi        string `json:"prodi"`
	SKNomor      string `json:"sk_nomor"`
	Periode      string `json:"periode"`
	Nominal      int64  `json:"nominal"`
	TanggalCair  string `json:"tanggal_cair"`
	NomorRekening string `json:"nomor_rekening"`
}

func (a *App) queryFindings(w http.ResponseWriter, r *http.Request) ([]findingRow, error) {
	var rows []findingRow
	q := a.db.Table("findings").
		Select("findings.*, students.nim, students.nama, students.prodi, disbursements.sk_nomor, disbursements.periode, disbursements.nominal, disbursements.tanggal_cair, accounts.nomor_rekening").
		Joins("JOIN disbursements ON disbursements.id = findings.disbursement_id").
		Joins("JOIN students ON students.id = disbursements.student_id").
		Joins("JOIN accounts ON accounts.id = disbursements.account_id")
	if s := r.URL.Query().Get("status"); s != "" {
		q = q.Where("findings.status = ?", s)
	}
	if b := r.URL.Query().Get("band"); b != "" {
		q = q.Where("findings.band = ?", b)
	}
	if err := q.Order("findings.risk_score desc").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (a *App) listFindings(w http.ResponseWriter, r *http.Request) {
	rows, err := a.queryFindings(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query findings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"findings": rows})
}

func (a *App) getFinding(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	var rows []findingRow
	if err := a.db.Table("findings").
		Select("findings.*, students.nim, students.nama, students.prodi, disbursements.sk_nomor, disbursements.periode, disbursements.nominal, disbursements.tanggal_cair, accounts.nomor_rekening").
		Joins("JOIN disbursements ON disbursements.id = findings.disbursement_id").
		Joins("JOIN students ON students.id = disbursements.student_id").
		Joins("JOIN accounts ON accounts.id = disbursements.account_id").
		Where("findings.id = ?", id).Scan(&rows).Error; err != nil || len(rows) == 0 {
		writeErr(w, http.StatusNotFound, "finding tidak ditemukan")
		return
	}
	var events []FindingEvent
	if err := a.db.Where("finding_id = ?", id).Order("at asc").Find(&events).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "query events")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"finding": rows[0], "events": events})
}

var allowedTransitions = map[string]map[string]bool{
	StatusBaru:         {StatusDiverifikasi: true, StatusSalahDeteksi: true},
	StatusDiverifikasi: {StatusSelesai: true, StatusSalahDeteksi: true},
	StatusSalahDeteksi: {StatusSelesai: true},
	StatusSelesai:      {},
}

func (a *App) setFindingStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad id")
		return
	}
	var in struct {
		ToStatus string `json:"to_status"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	var f Finding
	if err := a.db.First(&f, id).Error; err != nil {
		writeErr(w, http.StatusNotFound, "finding tidak ditemukan")
		return
	}
	if !allowedTransitions[f.Status][in.ToStatus] {
		writeErr(w, http.StatusConflict, "transisi tidak valid: "+f.Status+" → "+in.ToStatus)
		return
	}
	actor := userIDFrom(r)
	if actor == "" {
		actor = "admin"
	}
	// Capture before the update: GORM writes the new value back into the struct.
	from := f.Status
	ev := FindingEvent{FindingID: f.ID, FromStatus: from, ToStatus: in.ToStatus, Actor: actor, Note: in.Note, At: time.Now()}
	// Status and audit event move together or not at all (PRD §4.4).
	err = a.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Finding{}).Where("id = ?", f.ID).Update("status", in.ToStatus).Error; err != nil {
			return err
		}
		return tx.Create(&ev).Error
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "update status")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": in.ToStatus, "event": ev})
}

func (a *App) stats(w http.ResponseWriter, r *http.Request) {
	count := func(q string, args ...any) int64 {
		var n int64
		a.db.Raw(q, args...).Scan(&n)
		return n
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"records":           count("SELECT count(*) FROM disbursements"),
		"findings":          count("SELECT count(*) FROM findings"),
		"prioritas_tinggi":  count("SELECT count(*) FROM findings WHERE band = ?", BandPrioritas),
		"perlu_verifikasi":  count("SELECT count(*) FROM findings WHERE band = ?", BandVerifikasi),
		"rendah":            count("SELECT count(*) FROM findings WHERE band = ?", BandRendah),
		"baru":              count("SELECT count(*) FROM findings WHERE status = ?", StatusBaru),
		"salah_deteksi":     count("SELECT count(*) FROM findings WHERE status = ?", StatusSalahDeteksi),
	})
}

func (a *App) eval(w http.ResponseWriter, r *http.Request) {
	labels := generateSeed().Labels
	anomalous := map[string][]string{}
	for _, l := range labels {
		anomalous[l.NIM] = l.Rules
	}

	rows, err := a.queryFindings(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query findings")
		return
	}
	// queryFindings orders by risk desc — top-N slice is the precision@N input.
	top := rows
	if len(top) > 10 {
		top = top[:10]
	}
	hitsTop := 0
	for _, row := range top {
		if _, ok := anomalous[row.NIM]; ok {
			hitsTop++
		}
	}
	precisionAt10 := 0.0
	if len(top) > 0 {
		precisionAt10 = float64(hitsTop) / float64(len(top))
	}

	covered := 0
	foundByNIM := map[string]string{} // NIM -> finding.Rule
	for _, row := range rows {
		foundByNIM[row.NIM] = row.Rule
	}
	for _, l := range labels {
		got, ok := foundByNIM[l.NIM]
		if !ok {
			continue
		}
		all := true
		for _, want := range l.Rules {
			if !strings.Contains(got, want) {
				all = false
				break
			}
		}
		if all {
			covered++
		}
	}
	recall := float64(covered) / float64(len(labels))

	var cleanTotal int64
	a.db.Raw("SELECT count(*) FROM disbursements WHERE student_id NOT IN (SELECT id FROM students WHERE nim IN ?)", keys(anomalous)).Scan(&cleanTotal)
	fp := 0
	for _, row := range rows {
		if _, ok := anomalous[row.NIM]; !ok {
			fp++
		}
	}
	fpRate := 0.0
	if cleanTotal > 0 {
		fpRate = float64(fp) / float64(cleanTotal)
	}

	// Measured detection runtime on the current batch.
	var students []Student
	var accounts []Account
	var disbs []Disbursement
	a.db.Find(&students)
	a.db.Find(&accounts)
	a.db.Find(&disbs)
	start := time.Now()
	detect(students, accounts, disbs)
	runtimeMs := time.Since(start).Milliseconds()

	writeJSON(w, http.StatusOK, map[string]any{
		"precision_at_10":       precisionAt10,
		"recall":                recall,
		"planted":               len(labels),
		"covered":               covered,
		"false_positive_rate":   fpRate,
		"false_positives":       fp,
		"detection_runtime_ms":  runtimeMs,
		"methodology":           "sintetis berlabel (seed 42), bukan data asli",
	})
}

func keys(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
