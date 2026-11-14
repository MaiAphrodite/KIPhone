package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Spending rules over the partner-bank transaction feed (dummy data in this
// build). Four patterns per PRD extension:
//   T1 taruhan       - deposits to gambling merchants/agents
//   T2 drain         - most of the disbursement leaves within 7 days
//   T3 tunai         - cash withdrawal concentration (money leaves the rails)
//   T4 mewah         - luxury purchases above a threshold
//
// Same design contract as the disbursement rules: deterministic, auditable,
// evidence + a sentence a human can read, no automatic action.

const (
	txDrainShare      = 0.60 // T2 threshold of disbursement spent in 7 days
	txDrainDays       = 7
	txCashOutShare    = 0.45 // T3 threshold withdrawn in cash
	txLuxuryThreshold = 1_500_000
)

// txnBand uses spending-specific cutoffs: a single gambling hit is already
// "perlu verifikasi"; only combined patterns reach red.
func txnBand(risk float64) string {
	switch {
	case risk >= 0.75:
		return BandPrioritas
	case risk >= 0.30:
		return BandVerifikasi
	default:
		return BandRendah
	}
}

func detectSpending(students []Student, txns []Transaction) []SpendFinding {
	studentBy := map[uint]Student{}
	for _, s := range students {
		studentBy[s.ID] = s
	}
	byStudent := map[uint][]Transaction{}
	for _, t := range txns {
		byStudent[t.StudentID] = append(byStudent[t.StudentID], t)
	}

	var out []SpendFinding
	for studentID, list := range byStudent {
		s, ok := studentBy[studentID]
		if !ok {
			continue
		}
		if len(list) == 0 {
			continue
		}
		var first, last time.Time
		for i, t := range list {
			if i == 0 || t.Tanggal.Before(first) {
				first = t.Tanggal
			}
			if t.Tanggal.After(last) {
				last = t.Tanggal
			}
		}
		_ = last

		var total, taruhan, mewah, tunai, first7 int64
		var taruhanMerchants []string
		for _, t := range list {
			total += t.Nominal
			switch t.Kategori {
			case "taruhan":
				taruhan += t.Nominal
				if len(taruhanMerchants) < 4 && !containsStr(taruhanMerchants, t.Merchant) {
					taruhanMerchants = append(taruhanMerchants, fmt.Sprintf("%s (%s)", t.Merchant, rupiah(t.Nominal)))
				}
			case "mewah":
				mewah += t.Nominal
			case "tunai":
				tunai += t.Nominal
			}
			if t.Tanggal.Before(first.AddDate(0, 0, txDrainDays)) {
				first7 += t.Nominal
			}
		}

		const budget = int64(2_400_000)
		rules := []string{}
		evidence := map[string]any{}
		sentences := []string{}
		risk := 0.0

		if taruhan > 0 {
			rules = append(rules, "T1")
			risk += spendRuleWeights["T1"]
			evidence["T1"] = map[string]any{
				"total_taruhan": rupiah(taruhan),
				"transaksi":     len(taruhanMerchants),
				"merchant":      taruhanMerchants,
			}
			sentences = append(sentences, fmt.Sprintf("Ada %d transaksi ke merchant taruhan senilai %s.", len(taruhanMerchants), rupiah(taruhan)))
		}
		share := float64(first7) / float64(budget)
		if share >= txDrainShare {
			rules = append(rules, "T2")
			risk += spendRuleWeights["T2"]
			evidence["T2"] = map[string]any{
				"terpakai_7_hari": rupiah(first7),
				"persen_dana":     fmt.Sprintf("%.0f%%", share*100),
			}
			sentences = append(sentences, fmt.Sprintf("%.0f%% dana keluar dalam 7 hari pertama (%s dari %s).", share*100, rupiah(first7), rupiah(budget)))
		}
		{
			cashShare := float64(tunai) / float64(budget)
			if tunai > 0 && cashShare >= txCashOutShare {
				rules = append(rules, "T3")
				risk += spendRuleWeights["T3"]
				evidence["T3"] = map[string]any{
					"tarik_tunai": rupiah(tunai),
					"persen_dari_belanja": fmt.Sprintf("%.0f%%", cashShare*100),
				}
				sentences = append(sentences, fmt.Sprintf("%.0f%% pengeluaran berupa tarik tunai (%s).", cashShare*100, rupiah(tunai)))
			}
		}
		var luxury []string
		for _, t := range list {
			if t.Kategori == "mewah" && t.Nominal >= txLuxuryThreshold {
				luxury = append(luxury, fmt.Sprintf("%s (%s)", t.Merchant, rupiah(t.Nominal)))
			}
		}
		if len(luxury) > 0 {
			rules = append(rules, "T4")
			risk += spendRuleWeights["T4"]
			evidence["T4"] = map[string]any{
				"total_mewah": rupiah(mewah),
				"transaksi":   luxury,
			}
			sentences = append(sentences, fmt.Sprintf("Pembelian barang mewah di atas %s: %s.", rupiah(txLuxuryThreshold), strings.Join(luxury, ", ")))
		}

		if len(rules) == 0 {
			continue
		}
		if risk > 1 {
			risk = 1
		}
		ev, _ := json.Marshal(evidence)
		out = append(out, SpendFinding{
			StudentID:   s.ID,
			NIM:         s.NIM,
			Nama:        s.Nama,
			Rules:       strings.Join(rules, ","),
			RiskScore:   risk,
			Band:        txnBand(risk),
			Explanation: strings.Join(sentences, " "),
			Evidence:    string(ev),
			Status:      "baru",
			CreatedAt:   time.Now(),
		})
		_ = first
	}
	return out
}

func containsStr(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func rupiah(v int64) string {
	s := fmt.Sprintf("%d", v)
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += "."
		}
		out += string(c)
	}
	return "Rp " + out
}
