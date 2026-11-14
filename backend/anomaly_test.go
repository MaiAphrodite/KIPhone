package main

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

func testDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func testStudent(id uint, nim, nik, nama, status string, lastActive time.Time) Student {
	return Student{ID: id, NIM: nim, NIK: nik, Nama: nama, Prodi: "Informatika", StatusAkademik: status, LastActiveAt: lastActive}
}

func testAccount(id, studentID uint, rek, holder string, created time.Time) Account {
	return Account{ID: id, StudentID: studentID, NomorRekening: rek, NamaPemilik: holder, CreatedAt: created}
}

func testDisbursement(id, studentID, accountID uint, cair time.Time) Disbursement {
	return Disbursement{ID: id, StudentID: studentID, AccountID: accountID, SKNomor: "SK-TEST", Periode: "2025-1", Nominal: 2400000, TanggalCair: cair}
}

func findingsFor(findings []Finding, disbursementID uint) []Finding {
	var out []Finding
	for _, f := range findings {
		if f.DisbursementID == disbursementID {
			out = append(out, f)
		}
	}
	return out
}

func singleFinding(t *testing.T, findings []Finding, disbursementID uint) Finding {
	t.Helper()
	fs := findingsFor(findings, disbursementID)
	if len(fs) != 1 {
		t.Fatalf("disbursement %d: got %d findings, want exactly 1", disbursementID, len(fs))
	}
	return fs[0]
}

func hasRule(f Finding, rule string) bool {
	for _, r := range strings.Split(f.Rule, ",") {
		if r == rule {
			return true
		}
	}
	return false
}

func evidenceFor(t *testing.T, f Finding, rule string) map[string]any {
	t.Helper()
	var all map[string]map[string]any
	if err := json.Unmarshal([]byte(f.Evidence), &all); err != nil {
		t.Fatalf("finding %d evidence is not valid JSON: %v", f.ID, err)
	}
	ev, ok := all[rule]
	if !ok {
		t.Fatalf("finding %d evidence missing %s block: %s", f.ID, rule, f.Evidence)
	}
	return ev
}

func nearly(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestDetectR1DuplicateNIK(t *testing.T) {
	cair := testDate(2025, 8, 15)
	active := testDate(2025, 7, 1)

	t.Run("two students sharing a NIK fire R1 on both disbursements", func(t *testing.T) {
		students := []Student{
			testStudent(1, "NIM-1", "NIK-DUP", "Budi Santoso", "aktif", active),
			testStudent(2, "NIM-2", "NIK-DUP", "Siti Rahmawati", "aktif", active),
		}
		accounts := []Account{
			testAccount(1, 1, "REK-1", "Budi Santoso", testDate(2024, 9, 1)),
			testAccount(2, 2, "REK-2", "Siti Rahmawati", testDate(2024, 9, 1)),
		}
		disbs := []Disbursement{
			testDisbursement(1, 1, 1, cair),
			testDisbursement(2, 2, 2, cair),
		}

		findings := detect(students, accounts, disbs)
		if len(findings) != 2 {
			t.Fatalf("got %d findings, want 2", len(findings))
		}
		for _, id := range []uint{1, 2} {
			f := singleFinding(t, findings, id)
			if !hasRule(f, "R1") {
				t.Fatalf("disbursement %d rule = %q, want R1", id, f.Rule)
			}
			ev := evidenceFor(t, f, "R1")
			if got := ev["duplicate_count"].(float64); got != 2 {
				t.Fatalf("disbursement %d duplicate_count = %v, want 2", id, got)
			}
		}
	})

	t.Run("unique NIK does not fire R1", func(t *testing.T) {
		students := []Student{testStudent(1, "NIM-1", "NIK-1", "Budi Santoso", "aktif", active)}
		accounts := []Account{testAccount(1, 1, "REK-1", "Budi Santoso", testDate(2024, 9, 1))}
		disbs := []Disbursement{testDisbursement(1, 1, 1, cair)}

		if findings := detect(students, accounts, disbs); len(findings) != 0 {
			t.Fatalf("got %d findings for a clean record, want 0: %+v", len(findings), findings)
		}
	})
}

func TestDetectR2SharedAccount(t *testing.T) {
	cair := testDate(2025, 8, 15)
	active := testDate(2025, 7, 1)
	created := cair.AddDate(0, 0, -4)

	t.Run("shared rekening fires R2 with account age evidence", func(t *testing.T) {
		students := []Student{
			testStudent(1, "NIM-1", "NIK-1", "Budi Santoso", "aktif", active),
			testStudent(2, "NIM-2", "NIK-2", "Siti Rahmawati", "aktif", active),
		}
		accounts := []Account{
			testAccount(1, 1, "REK-SHARED", "Budi Santoso", created),
			testAccount(2, 2, "REK-SHARED", "Siti Rahmawati", created),
		}
		disbs := []Disbursement{
			testDisbursement(1, 1, 1, cair),
			testDisbursement(2, 2, 2, cair),
		}

		findings := detect(students, accounts, disbs)
		if len(findings) != 2 {
			t.Fatalf("got %d findings, want 2", len(findings))
		}
		for _, id := range []uint{1, 2} {
			f := singleFinding(t, findings, id)
			if !hasRule(f, "R2") {
				t.Fatalf("disbursement %d rule = %q, want R2", id, f.Rule)
			}
			ev := evidenceFor(t, f, "R2")
			age, ok := ev["rekening_umur_hari"].(float64)
			if !ok {
				t.Fatalf("disbursement %d R2 evidence missing rekening_umur_hari: %v", id, ev)
			}
			if age != 4 {
				t.Fatalf("disbursement %d rekening_umur_hari = %v, want 4", id, age)
			}
		}
	})

	t.Run("single account does not fire R2", func(t *testing.T) {
		students := []Student{testStudent(1, "NIM-1", "NIK-1", "Budi Santoso", "aktif", active)}
		accounts := []Account{testAccount(1, 1, "REK-SOLO", "Budi Santoso", created)}
		disbs := []Disbursement{testDisbursement(1, 1, 1, cair)}

		if findings := detect(students, accounts, disbs); len(findings) != 0 {
			t.Fatalf("got %d findings for a single-account record, want 0", len(findings))
		}
	})
}

func TestDetectR3NameMismatch(t *testing.T) {
	cair := testDate(2025, 8, 15)
	active := testDate(2025, 7, 1)

	tests := []struct {
		name    string
		student string
		holder  string
		wantR3  bool
	}{
		{name: "unrelated holder name fires R3", student: "Budi Santoso", holder: "Siti Rahmawati", wantR3: true},
		{name: "one-token surname typo does not fire R3", student: "Budi Santoso", holder: "Budi Santosa", wantR3: false},
		{name: "identical name does not fire R3", student: "Budi Santoso", holder: "Budi Santoso", wantR3: false},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			students := []Student{testStudent(1, "NIM-1", "NIK-1", tc.student, "aktif", active)}
			accounts := []Account{testAccount(1, 1, "REK-SOLO", tc.holder, testDate(2024, 9, 1))}
			disbs := []Disbursement{testDisbursement(1, 1, 1, cair)}

			findings := detect(students, accounts, disbs)
			if !tc.wantR3 {
				if len(findings) != 0 {
					t.Fatalf("case %d: got %d findings, want 0: %+v", i, len(findings), findings)
				}
				return
			}
			f := singleFinding(t, findings, 1)
			if !hasRule(f, "R3") {
				t.Fatalf("rule = %q, want R3", f.Rule)
			}
			ev := evidenceFor(t, f, "R3")
			raw, ok := ev["nama_match_score"].(string)
			if !ok {
				t.Fatalf("R3 evidence missing nama_match_score: %v", ev)
			}
			score, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				t.Fatalf("nama_match_score %q is not a number: %v", raw, err)
			}
			if score >= nameMatchThreshold {
				t.Fatalf("R3 fired with score %.4f, must be < %.2f", score, nameMatchThreshold)
			}
			if diff, ok := ev["token_berbeda"].(float64); !ok || diff < 1 {
				t.Fatalf("R3 evidence token_berbeda missing or zero: %v", ev)
			}
		})
	}
}

func TestDetectR4InactiveRecipient(t *testing.T) {
	cair := testDate(2025, 8, 15)

	tests := []struct {
		name       string
		status     string
		lastActive time.Time
		wantR4     bool
	}{
		{name: "cuti receiving after last active fires R4", status: "cuti", lastActive: testDate(2025, 1, 10), wantR4: true},
		{name: "lulus receiving after last active fires R4", status: "lulus", lastActive: testDate(2025, 1, 10), wantR4: true},
		{name: "nonaktif receiving after last active fires R4", status: "nonaktif", lastActive: testDate(2025, 1, 10), wantR4: true},
		{name: "aktif does not fire R4", status: "aktif", lastActive: testDate(2025, 1, 10), wantR4: false},
		{name: "inactive before disbursement does not fire R4", status: "cuti", lastActive: testDate(2025, 12, 1), wantR4: false},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			students := []Student{testStudent(1, "NIM-1", "NIK-1", "Budi Santoso", tc.status, tc.lastActive)}
			accounts := []Account{testAccount(1, 1, "REK-SOLO", "Budi Santoso", testDate(2024, 9, 1))}
			disbs := []Disbursement{testDisbursement(1, 1, 1, cair)}

			findings := detect(students, accounts, disbs)
			if !tc.wantR4 {
				if len(findings) != 0 {
					t.Fatalf("case %d: got %d findings, want 0: %+v", i, len(findings), findings)
				}
				return
			}
			f := singleFinding(t, findings, 1)
			if !hasRule(f, "R4") {
				t.Fatalf("rule = %q, want R4", f.Rule)
			}
			ev := evidenceFor(t, f, "R4")
			if got := ev["status_akademik"]; got != tc.status {
				t.Fatalf("status_akademik evidence = %v, want %s", got, tc.status)
			}
		})
	}
}

func TestDetectCompositeScoring(t *testing.T) {
	cair := testDate(2025, 8, 15)
	active := testDate(2025, 7, 1)
	old := testDate(2024, 9, 1)

	t.Run("R1+R2+R3 aggregates to 0.85 prioritas_tinggi in one finding", func(t *testing.T) {
		students := []Student{
			testStudent(1, "NIM-1", "NIK-DUP", "Budi Santoso", "aktif", active),
			testStudent(2, "NIM-2", "NIK-DUP", "Siti Rahmawati", "aktif", active),
		}
		accounts := []Account{
			testAccount(1, 1, "REK-SHARED", "Siti Rahmawati", old),
			testAccount(2, 2, "REK-SHARED", "Siti Rahmawati", old),
		}
		disbs := []Disbursement{
			testDisbursement(1, 1, 1, cair),
			testDisbursement(2, 2, 2, cair),
		}

		findings := detect(students, accounts, disbs)
		f := singleFinding(t, findings, 1)
		if f.Rule != "R1,R2,R3" {
			t.Fatalf("rule = %q, want R1,R2,R3", f.Rule)
		}
		if !nearly(f.RiskScore, 0.85) {
			t.Fatalf("risk = %.6f, want 0.85", f.RiskScore)
		}
		if f.Band != BandPrioritas {
			t.Fatalf("band = %q, want %q", f.Band, BandPrioritas)
		}
		for _, rule := range []string{"R1", "R2", "R3"} {
			evidenceFor(t, f, rule)
		}
	})

	t.Run("single R2 scores 0.30 rendah", func(t *testing.T) {
		students := []Student{
			testStudent(1, "NIM-1", "NIK-1", "Budi Santoso", "aktif", active),
			testStudent(2, "NIM-2", "NIK-2", "Siti Rahmawati", "aktif", active),
		}
		accounts := []Account{
			testAccount(1, 1, "REK-SHARED", "Budi Santoso", old),
			testAccount(2, 2, "REK-SHARED", "Siti Rahmawati", old),
		}
		disbs := []Disbursement{
			testDisbursement(1, 1, 1, cair),
			testDisbursement(2, 2, 2, cair),
		}

		findings := detect(students, accounts, disbs)
		f := singleFinding(t, findings, 1)
		if f.Rule != "R2" {
			t.Fatalf("rule = %q, want R2", f.Rule)
		}
		if !nearly(f.RiskScore, 0.30) {
			t.Fatalf("risk = %.6f, want 0.30", f.RiskScore)
		}
		if f.Band != BandRendah {
			t.Fatalf("band = %q, want %q", f.Band, BandRendah)
		}
	})
}

func TestBandThresholds(t *testing.T) {
	tests := []struct {
		risk float64
		want string
	}{
		{risk: 0, want: BandRendah},
		{risk: 0.30, want: BandRendah},
		{risk: 0.549, want: BandRendah},
		{risk: 0.55, want: BandVerifikasi},
		{risk: 0.79, want: BandVerifikasi},
		{risk: 0.80, want: BandPrioritas},
		{risk: 1, want: BandPrioritas},
	}

	for _, tc := range tests {
		if got := bandFor(tc.risk); got != tc.want {
			t.Fatalf("bandFor(%.3f) = %q, want %q", tc.risk, got, tc.want)
		}
	}
}
