package gerritchange

import "testing"

func TestParseChangeLists(t *testing.T) {
	r, err := ParseChangeListsSorted("chromium:123,chrome-internal:456")
	if err != nil {
		t.Error(err)
	}
	want1 := ChangeList{ShortHost: "chrome-internal", ChangeNum: 456}
	want2 := ChangeList{ShortHost: "chromium", ChangeNum: 123}
	if r[0] != want1 {
		t.Errorf("want %v, got %v", want1, r[0])
	}
	if r[1] != want2 {
		t.Errorf("want %v, got %v", want2, r[1])
	}
}

func TestParseChangeLists_BadInput(t *testing.T) {
	inputs := []string{
		"Android:123",           // capital letter
		"dog999",                // no colon
		"cat:",                  // no change num
		":123",                  // no short host
		"chromium:1;chromium:2", // uses semi-colon rather than comma

	}
	for _, i := range inputs {
		_, err := ParseChangeListsSorted(i)
		if err == nil {
			t.Errorf("expected an error, but got none for input %v", i)
		}
	}
}
