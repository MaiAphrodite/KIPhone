package main

import "strings"

// jaroWinkler returns the Jaro-Winkler similarity of two names in [0,1].
// Used by R3 to compare the SK name against the account holder name.
func jaroWinkler(a, b string) float64 {
	ar, br := []rune(strings.ToLower(strings.TrimSpace(a))), []rune(strings.ToLower(strings.TrimSpace(b)))
	la, lb := len(ar), len(br)
	if la == 0 || lb == 0 {
		return 0
	}
	matchDist := max(la, lb)/2 - 1
	if matchDist < 0 {
		matchDist = 0
	}
	aMatch := make([]bool, la)
	bMatch := make([]bool, lb)
	matches := 0
	for i := 0; i < la; i++ {
		start := max(0, i-matchDist)
		end := min(i+matchDist+1, lb)
		for j := start; j < end; j++ {
			if bMatch[j] || ar[i] != br[j] {
				continue
			}
			aMatch[i], bMatch[j] = true, true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}
	transpositions := 0
	k := 0
	for i := 0; i < la; i++ {
		if !aMatch[i] {
			continue
		}
		for !bMatch[k] {
			k++
		}
		if ar[i] != br[k] {
			transpositions++
		}
		k++
	}
	transpositions /= 2
	m := float64(matches)
	jaro := (m/float64(la) + m/float64(lb) + (m-float64(transpositions))/m) / 3
	prefix := 0
	for prefix < 4 && prefix < la && prefix < lb && ar[prefix] == br[prefix] {
		prefix++
	}
	return jaro + float64(prefix)*0.1*(1-jaro)
}

// tokenDiff counts tokens present in exactly one of the two names
// (symmetric difference of token sets).
func tokenDiff(a, b string) int {
	set := func(s string) map[string]bool {
		out := map[string]bool{}
		for _, t := range strings.Fields(strings.ToLower(s)) {
			out[t] = true
		}
		return out
	}
	sa, sb := set(a), set(b)
	diff := 0
	for t := range sa {
		if !sb[t] {
			diff++
		}
	}
	for t := range sb {
		if !sa[t] {
			diff++
		}
	}
	return diff
}
