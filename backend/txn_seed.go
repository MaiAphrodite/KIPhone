package main

import (
	"math/rand"
	"time"
)

// Dummy transaction feed from the partner bank. 215 recipients, ~40
// transactions each over the 30 days after disbursement, with 14 planted
// misuse cases across four patterns. Deterministic (seed 42) so the demo
// reproduces. The disbursement amount is Rp 2.400.000.

const txnSeedValue = 42

type txnSeed struct {
	Transactions []Transaction
	Labels       []seedLabel // NIM + rules, reused for /api/tx/eval
}

var (
	makanMerchants = []string{"WARTEG BAHARI", "RM PADANG SEDERHANA", "MIE GACOAN", "KOPI KENANGAN", "INDOMARET", "ALFAMART"}
	transportM     = []string{"GRAB BIKE", "GOJEK", "KRL COMMUTER", "TRANSJAKARTA", "PERTAMINA"}
	pulsaM         = []string{"TELKOMSEL", "INDOSAT", "XL AXIATA", "TOKOPEDIA PULSA"}
	belanjaM       = []string{"SHOPEE", "TOKOPEDIA", "LAZADA", "UNIQLO", "SOCIOLLA"}
	hiburanM       = []string{"CGV CINEMA", "XXI", "SPOTIFY", "STEAM", "MLBB DIAMOND"}
	mewahM         = []string{"TOKO EMAS MULIA", "BUTIK HERMES", "APPLE STORE", "HOTEL BINTANG 5", "GOLF CLUB"}
	taruhanM       = []string{"ZEUS SLOT", "TOGEL HK", "DEPOSIT AGEN 88", "GACOR MAXWIN", "SLOT ONLINE 4D", "BANDAR TARUHAN"}
)

func generateTxns(students []Student, accounts []Account, disbs []Disbursement) txnSeed {
	r := rand.New(rand.NewSource(txnSeedValue))
	out := txnSeed{}

	acctBy := map[uint]Account{}
	for _, a := range accounts {
		acctBy[a.StudentID] = a
	}
	disbBy := map[uint]Disbursement{}
	for _, d := range disbs {
		disbBy[d.StudentID] = d
	}

	pick := func(list []string) string { return list[r.Intn(len(list))] }
	amt := func(lo, hi int64) int64 { return lo + r.Int63n(hi-lo) }

	// planted misuse, keyed by student index
	type plant struct {
		rules []string
		kind  string // taruhan | mewah | drain | tunai
	}
	plants := map[int]plant{}
	plan := []plant{
		{[]string{"T1"}, "taruhan"},
		{[]string{"T1"}, "taruhan"},
		{[]string{"T1", "T2"}, "taruhan"},
		{[]string{"T1", "T2"}, "taruhan"},
		{[]string{"T1", "T3"}, "taruhan"},
		{[]string{"T4"}, "mewah"},
		{[]string{"T4"}, "mewah"},
		{[]string{"T4"}, "mewah"},
		{[]string{"T2"}, "drain"},
		{[]string{"T2"}, "drain"},
		{[]string{"T2"}, "drain"},
		{[]string{"T3"}, "tunai"},
		{[]string{"T3"}, "tunai"},
		{[]string{"T2", "T3"}, "tunai"},
	}
	// spread the planted cases across the roster instead of the first 14 rows
	step := len(students) / len(plan)
	if step < 1 {
		step = 1
	}
	for i, p := range plan {
		idx := (i * step) % len(students)
		plants[idx] = p
	}

	txID := uint(1)
	add := func(s Student, t time.Time, merchant, kategori string, nominal int64, channel string) {
		out.Transactions = append(out.Transactions, Transaction{
			ID: txID, StudentID: s.ID, Tanggal: t, Merchant: merchant,
			Kategori: kategori, Nominal: nominal, Arah: "keluar", Channel: channel,
		})
		txID++
	}

	for i, s := range students {
		d, ok := disbBy[s.ID]
		if !ok {
			continue
		}
		start := d.TanggalCair.AddDate(0, 0, 1)
		p := plants[i]
		spent := int64(0)
		budget := int64(2_400_000)

		// normal daily life: ~36 transactions over 30 days
		for n := 0; n < 36; n++ {
			day := start.AddDate(0, 0, r.Intn(30))
			switch r.Intn(10) {
			case 0, 1, 2, 3:
				v := amt(15_000, 45_000)
				add(s, day, pick(makanMerchants), "makan", v, "qris")
				spent += v
			case 4, 5:
				v := amt(10_000, 35_000)
				add(s, day, pick(transportM), "transport", v, "qris")
				spent += v
			case 6:
				v := amt(20_000, 80_000)
				add(s, day, pick(pulsaM), "pulsa", v, "ewallet")
				spent += v
			case 7, 8:
				v := amt(50_000, 250_000)
				add(s, day, pick(belanjaM), "belanja", v, "ewallet")
				spent += v
			case 9:
				v := amt(40_000, 120_000)
				add(s, day, pick(hiburanM), "hiburan", v, "qris")
				spent += v
			}
		}

		switch p.kind {
		case "taruhan":
			// repeated deposits to gambling merchants in the first week
			for n := 0; n < 6; n++ {
				v := amt(100_000, 400_000)
				add(s, start.AddDate(0, 0, r.Intn(7)), pick(taruhanM), "taruhan", v, "ewallet")
				spent += v
			}
		case "mewah":
			for n := 0; n < 2; n++ {
				v := amt(1_500_000, 2_200_000)
				add(s, start.AddDate(0, 0, r.Intn(10)), pick(mewahM), "mewah", v, "debit")
				spent += v
			}
		case "drain":
			// large transfers out within the first 5 days
			for n := 0; n < 3; n++ {
				v := amt(600_000, 900_000)
				add(s, start.AddDate(0, 0, r.Intn(5)), "TRANSFER KELUAR", "transfer", v, "transfer")
				spent += v
			}
		case "tunai":
			for n := 0; n < 2; n++ {
				v := amt(700_000, 1_000_000)
				add(s, start.AddDate(0, 0, r.Intn(12)), "TARIK TUNAI ATM", "tunai", v, "atm")
				spent += v
			}
		}

		if p.kind != "" {
			out.Labels = append(out.Labels, seedLabel{NIM: s.NIM, Rules: p.rules})
		}
		_ = spent
		_ = budget
	}
	return out
}
