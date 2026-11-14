package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Synthetic batch per PRD §6: 215 students (200 clean + 15 planted anomalies),
// each with one account and one disbursement. Deterministic under seed 42 so
// the demo reproduces on any machine. Labels are regenerated for /api/eval.

const seedValue = 42

type seedLabel struct {
	NIM   string   `json:"nim"`
	Rules []string `json:"rules"`
}

type seedData struct {
	Students      []Student
	Accounts      []Account
	Disbursements []Disbursement
	Labels        []seedLabel
}

var firstNames = []string{
	"Ahmad", "Budi", "Citra", "Dewi", "Eko", "Fitri", "Gilang", "Hana",
	"Irfan", "Joko", "Kartika", "Lestari", "Made", "Nadia", "Rizki", "Sari",
	"Taufik", "Umar", "Wulan", "Yusuf", "Agus", "Ratna", "Dedi", "Sinta",
}
var lastNames = []string{
	"Santoso", "Wijaya", "Pratama", "Kusuma", "Nugroho", "Saputra",
	"Rahmawati", "Putri", "Hidayat", "Siregar", "Lubis", "Hakim",
	"Maulana", "Fauzi", "Anggraini", "Wibowo",
}
var prodis = []string{
	"Informatika", "Sistem Informasi", "Teknik Elektro", "Manajemen",
	"Akuntansi", "Hukum", "Psikologi", "Teknik Sipil",
}

func generateSeed() seedData {
	r := rand.New(rand.NewSource(seedValue))
	cair := time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)

	d := seedData{}
	usedNIK := map[string]bool{}
	usedNIM := map[string]bool{}
	usedRek := map[string]bool{}

	digits := func(n int) string {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte('0' + r.Intn(10))
		}
		return string(b)
	}
	uniq := func(n int, used map[string]bool) string {
		for {
			s := digits(n)
			if !used[s] {
				used[s] = true
				return s
			}
		}
	}
	name := func() string { return firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))] }

	var nextID uint = 1
	addStudent := func(nama, nik, status string, lastActive time.Time, rekOwner string, rekCreated time.Time, periode string) (Student, Account, Disbursement) {
		s := Student{ID: nextID, NIM: uniq(10, usedNIM), NIK: nik, Nama: nama, Prodi: prodis[r.Intn(len(prodis))], StatusAkademik: status, LastActiveAt: lastActive}
		a := Account{ID: nextID, StudentID: s.ID, NomorRekening: uniq(12, usedRek), NamaPemilik: rekOwner, CreatedAt: rekCreated}
		db := Disbursement{ID: nextID, StudentID: s.ID, AccountID: a.ID, SKNomor: fmt.Sprintf("SK-%s-%04d", periode, nextID), Periode: periode, Nominal: 2400000, TanggalCair: cair}
		nextID++
		d.Students = append(d.Students, s)
		d.Accounts = append(d.Accounts, a)
		d.Disbursements = append(d.Disbursements, db)
		return s, a, db
	}

	active := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	oldAccount := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)

	// ── Planted anomalies (labels regenerated for /api/eval) ───────────

	// R1 ×4: two NIK pairs claimed by two different students each.
	// The first pair is the headline case: both share one rekening (R2) and the
	// second claimer's account holder name is unrelated (R3) → red band.
	for p := 0; p < 2; p++ {
		nik := uniq(16, usedNIK)
		nm1 := name()
		s1, a1, _ := addStudent(nm1, nik, "aktif", active, nm1, oldAccount, "2024-2")
		s2nama := name()
		s2 := Student{ID: nextID, NIM: uniq(10, usedNIM), NIK: nik, Nama: s2nama, Prodi: prodis[r.Intn(len(prodis))], StatusAkademik: "aktif", LastActiveAt: active}
		a2 := Account{ID: nextID, StudentID: s2.ID, NomorRekening: uniq(12, usedRek), NamaPemilik: s2nama, CreatedAt: oldAccount}
		db2 := Disbursement{ID: nextID, StudentID: s2.ID, AccountID: a2.ID, SKNomor: fmt.Sprintf("SK-2025-1-%04d", nextID), Periode: "2025-1", Nominal: 2400000, TanggalCair: cair}
		nextID++
		s1Rules := []string{"R1"}
		s2Rules := []string{"R1"}
		if p == 0 {
			// shared rekening between the pair + unrelated holder name on s2
			a2.NomorRekening = a1.NomorRekening
			delete(usedRek, a2.NomorRekening)
			a2.NamaPemilik = firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))]
			for a2.NamaPemilik == s2nama || a2.NamaPemilik == nm1 {
				a2.NamaPemilik = firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))]
			}
			s1Rules = []string{"R1", "R2"}
			s2Rules = []string{"R1", "R2", "R3"}
		}
		d.Students = append(d.Students, s2)
		d.Accounts = append(d.Accounts, a2)
		d.Disbursements = append(d.Disbursements, db2)
		d.Labels = append(d.Labels,
			seedLabel{NIM: s1.NIM, Rules: s1Rules},
			seedLabel{NIM: s2.NIM, Rules: s2Rules},
		)
	}

	// R2 ×4: two accounts each shared by two students; accounts created
	// 4 days before disbursement (young-account signal).
	for p := 0; p < 2; p++ {
		rek := uniq(12, usedRek)
		rekCreated := cair.AddDate(0, 0, -4)
		nama1 := name()
		s1, a1, _ := addStudent(nama1, uniq(16, usedNIK), "aktif", active, nama1, rekCreated, "2025-1")
		nama2 := name()
		s2, a2, _ := addStudent(nama2, uniq(16, usedNIK), "aktif", active, nama2, rekCreated, "2025-1")
		// rewire both accounts to the shared number
		d.Accounts[a1.ID-1].NomorRekening = rek
		d.Accounts[a2.ID-1].NomorRekening = rek
		usedRek[rek] = true
		// one shared account also mismatches the holder name (R2+R3, red band)
		if p == 0 {
			d.Accounts[a2.ID-1].NamaPemilik = firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))]
			d.Labels = append(d.Labels, seedLabel{NIM: s2.NIM, Rules: []string{"R2", "R3"}})
		} else {
			d.Labels = append(d.Labels, seedLabel{NIM: s2.NIM, Rules: []string{"R2"}})
		}
		d.Labels = append(d.Labels, seedLabel{NIM: s1.NIM, Rules: []string{"R2"}})
	}

	// R3 ×4: account holder name unrelated to the SK name.
	for p := 0; p < 4; p++ {
		nm := name()
		other := firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))]
		for other == nm {
			other = firstNames[r.Intn(len(firstNames))] + " " + lastNames[r.Intn(len(lastNames))]
		}
		s, _, _ := addStudent(nm, uniq(16, usedNIK), "aktif", active, other, oldAccount, "2025-1")
		d.Labels = append(d.Labels, seedLabel{NIM: s.NIM, Rules: []string{"R3"}})
	}

	// R4 ×3: inactive/graduated students still receiving funds.
	statuses := []string{"lulus", "cuti", "nonaktif"}
	for p := 0; p < 3; p++ {
		nm := name()
		lastActive := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		s, _, _ := addStudent(nm, uniq(16, usedNIK), statuses[p], lastActive, nm, oldAccount, "2025-1")
		d.Labels = append(d.Labels, seedLabel{NIM: s.NIM, Rules: []string{"R4"}})
	}

	// ── 200 clean records ───────────────────────────────────────────────
	for len(d.Students) < 215 {
		nm := name()
		addStudent(nm, uniq(16, usedNIK), "aktif", active, nm, oldAccount.AddDate(0, 0, -r.Intn(120)), "2025-1")
	}

	return d
}
