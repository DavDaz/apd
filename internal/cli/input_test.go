package cli

import (
	"strings"
	"testing"
)

func TestReadIntentCommands(t *testing.T) {
	cases := []struct {
		input string
		want  IntentKind
	}{
		{"/help\n", IntentHelp},
		{"/skip\n", IntentSkip},
		{"/back\n", IntentBack},
		{"/done\n", IntentDone},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := NewInput(strings.NewReader(tc.input)).ReadIntent()
			if err != nil {
				t.Fatalf("ReadIntent() error = %v", err)
			}
			if got.Kind != tc.want {
				t.Fatalf("ReadIntent().Kind = %s, want %s", got.Kind, tc.want)
			}
		})
	}
}

func TestReadIntentAnswers(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "hello\n\n", "hello"},
		{"multi-line", "hello\nworld\n\n", "hello\nworld"},
		{"command-like inside answer", "first\n/done\n\n", "first\n/done"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewInput(strings.NewReader(tc.input)).ReadIntent()
			if err != nil {
				t.Fatalf("ReadIntent() error = %v", err)
			}
			if got.Kind != IntentAnswer || got.Answer != tc.want {
				t.Fatalf("ReadIntent() = %+v, want answer %q", got, tc.want)
			}
		})
	}
}

func TestSelectType(t *testing.T) {
	t.Run("numeric choice needs no extra blank line", func(t *testing.T) {
		got, err := SelectType(strings.NewReader("2\n"), &strings.Builder{}, []string{"bug", "product"})
		if err != nil {
			t.Fatalf("SelectType() error = %v", err)
		}
		if got != "product" {
			t.Fatalf("SelectType() = %q, want product", got)
		}
	})
	t.Run("id choice needs no extra blank line", func(t *testing.T) {
		got, err := SelectType(strings.NewReader("product\n"), &strings.Builder{}, []string{"bug", "product"})
		if err != nil {
			t.Fatalf("SelectType() error = %v", err)
		}
		if got != "product" {
			t.Fatalf("SelectType() = %q, want product", got)
		}
	})
}
