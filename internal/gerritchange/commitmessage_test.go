package gerritchange

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestCommitMessage_WithFooter_ReplaceExisting(t *testing.T) {
	input := `Fix that bug

My-Footer: my old value
Change-Id: abcdef
`
	want := CommitMessage{Message: `Fix that bug

My-Footer: my new value
Change-Id: abcdef
`}
	got, err := CommitMessage{Message: input}.WithFooter("My-Footer", "my new value")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("got %v,\nwant %v\ndiff %v", got, want, cmp.Diff(got, want))
	}
}

func TestCommitMessage_WithFooter_AddNew(t *testing.T) {
	input := `Fix that bug

Change-Id: abcdef
`
	want := CommitMessage{Message: `Fix that bug

My-Footer: my new value
Change-Id: abcdef
`}
	got, err := CommitMessage{Message: input}.WithFooter("My-Footer", "my new value")
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("got %v,\nwant %v\ndiff %v", got, want, cmp.Diff(got, want))
	}
}

func TestCommitMessage_WithFooter_FailsIfFooterFoundTwice(t *testing.T) {
	input := `Fix that bug

My-Footer: value 1
My-Footer: value 2
Change-Id: abcdef
`
	_, err := CommitMessage{Message: input}.WithFooter("My-Footer", "my new value")
	if err == nil {
		t.Errorf("expected error, instead nil")
	}
}
