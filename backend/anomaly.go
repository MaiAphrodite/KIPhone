package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Detection per PRD §4. One finding per disbursement aggregating every rule
// that hit it; risk is the composite of the hit-rule weights.

var ruleWeights = map[string]float64{"R1": 0.30, "R2": 0.30, "R3": 0.25, "R4": 0.15}

const (
	BandRendah     = "rendah"
	BandVerifikasi = "perlu_verifikasi"
	BandPrioritas  = "prioritas_tinggi"

	StatusBaru          = "baru"
	StatusDiverifikasi  = "diverifikasi"
	StatusSalahDeteksi  = "salah_deteksi"
	StatusSelesai       = "selesai"
	nameMatchThreshold  = 0.85
	youngAccountMaxDays = 30
)

func bandFor(risk float64) string {
	switch {
	case risk >= 0.80:
		return BandPrioritas
	case risk >= 0.55:
		return BandVerifikasi
	default:
		return BandRendah
	}
}

type ruleHit struct {
	rule      string
	evidence  map[string]any
	sentence  string
}

// detect runs R1–R4 over the loaded batch and returns one Finding per
// affected disbursement (unsaved; the handler persists them).
func detect(students []Student, accounts []Account, disbs []Disbursement) []Finding {
	studentByID := make(map[uint]Student, len(students))
	for _, s := range students {
		studentByID[s.ID] = s
	}
	accountByID := make(map[uint]Account, len(accounts))
	for _, a := range accounts {
		accountByID[a.ID] = a
	}

	hitsByDisb := map[uint][]ruleHit{}
	add := func(d Disbursement, h ruleHit) {
		hitsByDisb[d.ID] = append(hitsByDisb[d.ID], h)
	}

	// R1 — penerima ganda: the same NIK *or* NIM claimed by more than one
	// student (PRD §4.1 says "NIK atau NIM").
	dupStudent := map[uint][]Student{}
	mergeDuplicates := func(keyOf func(Student) string) {
		groups := map[string][]Student{}
		for _, s := range students {
			k := keyOf(s)
			if k == "" {
				continue
			}
			groups[k] = append(groups[k], s)
		}
		for _, group := range groups {
			if len(group) < 2 {
				continue
			}
			for _, s := range group {
				seen := map[uint]bool{}
				for _, existing := range dupStudent[s.ID] {
					seen[existing.ID] = true
				}
				for _, o := range group {
					if o.ID != s.ID && !seen[o.ID] {
						dupStudent[s.ID] = append(dupStudent[s.ID], o)
					}
				}
			}
		}
	}
	mergeDuplicates(func(s Student) string { return s.NIK })
	mergeDuplicates(func(s Student) string { return s.NIM })

	// R2 — rekening bersama: >1 student on one account number.
	ownerByRek := map[string][]Account{}
	for _, a := range accounts {
		ownerByRek[a.NomorRekening] = append(ownerByRek[a.NomorRekening], a)
	}
	sharedStudents := map[uint][]uint{} // student -> other students on same account
	for _, list := range ownerByRek {
		studentIDs := map[uint]bool{}
		for _, a := range list {
			studentIDs[a.StudentID] = true
		}
		if len(studentIDs) < 2 {
			continue
		}
		for _, a := range list {
			var others []uint
			for _, b := range list {
				if b.StudentID != a.StudentID {
					others = append(others, b.StudentID)
				}
			}
			sharedStudents[a.StudentID] = others
		}
	}

	for _, d := range disbs {
		s := studentByID[d.StudentID]
		a := accountByID[d.AccountID]

		if others, ok := dupStudent[s.ID]; ok {
			names := make([]string, 0, len(others))
			for _, o := range others {
				names = append(names, fmt.Sprintf("%s (%s)", o.Nama, o.NIM))
			}
			add(d, ruleHit{
				rule: "R1",
				evidence: map[string]any{
					"nik":            s.NIK,
					"claimed_by":     append([]string{fmt.Sprintf("%s (%s)", s.Nama, s.NIM)}, names...),
					"duplicate_count": len(others) + 1,
				},
				sentence: fmt.Sprintf("NIK %s dipakai oleh %d mahasiswa: %s.", s.NIK, len(others)+1, strings.Join(names, ", ")),
			})
		}

		if others, ok := sharedStudents[s.ID]; ok {
			names := make([]string, 0, len(others))
			for _, oid := range others {
				o := studentByID[oid]
				names = append(names, fmt.Sprintf("%s (%s)", o.Nama, o.NIM))
			}
			ageDays := int(d.TanggalCair.Sub(a.CreatedAt).Hours() / 24)
			sent := fmt.Sprintf("Rekening %s dipakai oleh %d mahasiswa: %s.", a.NomorRekening, len(others)+1, strings.Join(names, ", "))
			if ageDays >= 0 && ageDays <= youngAccountMaxDays {
				sent += fmt.Sprintf(" Rekening baru berumur %d hari saat pencairan.", ageDays)
			}
			add(d, ruleHit{
				rule: "R2",
				evidence: map[string]any{
					"nomor_rekening":   a.NomorRekening,
					"pemilik_rekening": a.NamaPemilik,
					"dipakai_oleh":     append([]string{fmt.Sprintf("%s (%s)", s.Nama, s.NIM)}, names...),
					"rekening_umur_hari": ageDays,
				},
				sentence: sent,
			})
		}

		score := jaroWinkler(s.Nama, a.NamaPemilik)
		if score < nameMatchThreshold {
			diff := tokenDiff(s.Nama, a.NamaPemilik)
			add(d, ruleHit{
				rule: "R3",
				evidence: map[string]any{
					"nama_sk":           s.Nama,
					"nama_rekening":     a.NamaPemilik,
					"nama_match_score":  fmt.Sprintf("%.2f", score),
					"token_berbeda":     diff,
				},
				sentence: fmt.Sprintf("Nama rekening (%s) berbeda %d token dari nama SK (%s); skor kemiripan %.2f.", a.NamaPemilik, diff, s.Nama, score),
			})
		}

		if s.StatusAkademik != "aktif" && d.TanggalCair.After(s.LastActiveAt) {
			add(d, ruleHit{
				rule: "R4",
				evidence: map[string]any{
					"status_akademik": s.StatusAkademik,
					"last_active_at":  s.LastActiveAt.Format("2006-01-02"),
					"tanggal_cair":    d.TanggalCair.Format("2006-01-02"),
				},
				sentence: fmt.Sprintf("Status akademik %s sejak %s, tetapi dana cair %s.", s.StatusAkademik, s.LastActiveAt.Format("2006-01-02"), d.TanggalCair.Format("2006-01-02")),
			})
		}
	}

	findings := make([]Finding, 0, len(hitsByDisb))
	for disbID, hits := range hitsByDisb {
		rules := make([]string, 0, len(hits))
		evidence := map[string]any{}
		sentences := make([]string, 0, len(hits))
		risk := 0.0
		for _, h := range hits {
			rules = append(rules, h.rule)
			evidence[h.rule] = h.evidence
			sentences = append(sentences, h.sentence)
			risk += ruleWeights[h.rule]
		}
		if risk > 1 {
			risk = 1
		}
		evidenceJSON, _ := json.Marshal(evidence)
		findings = append(findings, Finding{
			DisbursementID: disbID,
			Rule:           strings.Join(rules, ","),
			RiskScore:      risk,
			Band:           bandFor(risk),
			Explanation:    strings.Join(sentences, " "),
			Evidence:       string(evidenceJSON),
			Status:         StatusBaru,
			CreatedAt:      time.Now(),
		})
	}
	return findings
}
