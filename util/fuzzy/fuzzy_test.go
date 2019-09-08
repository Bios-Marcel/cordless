package fuzzy

import (
	"testing"
)

func TestFuzzyScore(t *testing.T) {
	scoreA := Score("Marc", "Marcel#3414")
	scoreB := Score("Marc", "Maurice")
	if scoreA <= scoreB {
		t.Errorf("Incorrect score rating")
	}

	scoreC := Score("Marc", "Marchesi#9331")
	if scoreC <= scoreB {
		t.Errorf("Incorrect score rating")
	}

	arr := []string{"tests", "test", "Testosterone", "atesta", "bob"}
	sorted := SortSearchResults(ScoreSearch("Te", arr))
	expected := [4]string{"Testosterone", "test", "tests", "atesta"}

	if len(sorted) != len(expected) {
		t.Errorf("Expected length of %d, but received %d.\n", len(expected), len(sorted))
	}

	var results [4]string
	for i, result := range sorted {
		results[i] = result.Key
	}

	if results != expected {
		t.Errorf("Expected\n%v\nbut received\n%v\n", expected, results)
	}

	if Score("S", "cvcx") >= 0 {
		t.Errorf("Expected score to be < 0\n")
	}

	if Score("S", "scheme god#2313") <= 0 {
		t.Errorf("Expected score to be > 0\n")
	}

	if Score("K", "KC") <= 0 {
		t.Errorf("Expected score to be > 0\n")
	}

	if Score("s", "space") <= 0 {
		t.Errorf("Expected score to be > 0\n")
	}

	scoreA = Score("sout", "lexjusto#4214")
	scoreB = Score("sout", "Southclaws#4153")

	if scoreA >= scoreB {
		t.Errorf("Unexpected score difference.\n")
	}

	if Score("Hual#1299 ", "Thur_MaliGnY#1112") >= 0 {
		t.Errorf("Unexpected non-negative score")
	}

}
