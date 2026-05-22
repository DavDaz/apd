// Package cli handles terminal rendering and guided input parsing.
package cli

// IntentKind identifies one user action from guided input.
type IntentKind string

const (
	// IntentAnswer submits answer text for the current section.
	IntentAnswer IntentKind = "answer"
	// IntentHelp requests contextual help for the current section.
	IntentHelp IntentKind = "help"
	// IntentSkip skips the current section.
	IntentSkip IntentKind = "skip"
	// IntentBack returns to the previous section.
	IntentBack IntentKind = "back"
	// IntentDone completes the workflow early.
	IntentDone IntentKind = "done"
)

// Intent is one parsed user action.
type Intent struct {
	Kind   IntentKind
	Answer string
}

func commandIntent(line string) (Intent, bool) {
	switch line {
	case "/help":
		return Intent{Kind: IntentHelp}, true
	case "/skip":
		return Intent{Kind: IntentSkip}, true
	case "/back":
		return Intent{Kind: IntentBack}, true
	case "/done":
		return Intent{Kind: IntentDone}, true
	default:
		return Intent{}, false
	}
}
