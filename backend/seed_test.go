package main

import "testing"

func TestGenerateSeedCountsAndLabels(t *testing.T) {
	seed := generateSeed()

	if got := len(seed.Students); got != 215 {
		t.Fatalf("students = %d, want 215", got)
	}
	if got := len(seed.Accounts); got != 215 {
		t.Fatalf("accounts = %d, want 215", got)
	}
	if got := len(seed.Disbursements); got != 215 {
		t.Fatalf("disbursements = %d, want 215", got)
	}
	if got := len(seed.Labels); got != 15 {
		t.Fatalf("labels = %d, want 15", got)
	}

	nims := make(map[string]bool, len(seed.Students))
	for _, s := range seed.Students {
		nims[s.NIM] = true
	}
	for _, label := range seed.Labels {
		if label.NIM == "" {
			t.Fatalf("label with empty NIM: %+v", label)
		}
		if !nims[label.NIM] {
			t.Fatalf("label NIM %s does not exist in students", label.NIM)
		}
		if len(label.Rules) == 0 {
			t.Fatalf("label %s has no rules", label.NIM)
		}
	}
}

func TestGenerateSeedIsDeterministic(t *testing.T) {
	a := generateSeed()
	b := generateSeed()

	if len(a.Students) != len(b.Students) || len(a.Labels) != len(b.Labels) {
		t.Fatalf("seed shape is not deterministic: students %d/%d labels %d/%d",
			len(a.Students), len(b.Students), len(a.Labels), len(b.Labels))
	}
	for i := range a.Students {
		if a.Students[i] != b.Students[i] {
			t.Fatalf("student %d differs between runs: %+v vs %+v", i, a.Students[i], b.Students[i])
		}
	}
	for i := range a.Labels {
		if a.Labels[i].NIM != b.Labels[i].NIM {
			t.Fatalf("label %d NIM differs between runs: %s vs %s", i, a.Labels[i].NIM, b.Labels[i].NIM)
		}
	}
}

func TestDetectCoversAllPlantedLabels(t *testing.T) {
	seed := generateSeed()
	findings := detect(seed.Students, seed.Accounts, seed.Disbursements)

	if len(findings) < 15 {
		t.Fatalf("detect produced %d findings, want >= 15", len(findings))
	}

	studentNIM := make(map[uint]string, len(seed.Students))
	for _, s := range seed.Students {
		studentNIM[s.ID] = s.NIM
	}
	disbNIM := make(map[uint]string, len(seed.Disbursements))
	for _, d := range seed.Disbursements {
		disbNIM[d.ID] = studentNIM[d.StudentID]
	}

	found := make(map[string]bool, len(findings))
	for _, f := range findings {
		if f.Status != StatusBaru {
			t.Fatalf("finding for disbursement %d has status %q, want %q", f.DisbursementID, f.Status, StatusBaru)
		}
		nis := disbNIM[f.DisbursementID]
		if nis == "" {
			t.Fatalf("finding references unknown disbursement %d", f.DisbursementID)
		}
		found[nis] = true
	}

	for _, label := range seed.Labels {
		if !found[label.NIM] {
			t.Fatalf("planted label NIM %s (%v) has no finding", label.NIM, label.Rules)
		}
	}

	if len(findings) > 15 {
		t.Logf("detect produced %d findings for 15 labels; %d extra may be false positives", len(findings), len(findings)-15)
	}
}
