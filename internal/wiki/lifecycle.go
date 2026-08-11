package wiki

// NextAction returns the deterministic, non-semantic next safe action.
func NextAction(status Status) string {
	switch status {
	case StatusInitialized:
		return "Register a local source."
	case StatusRegistered:
		return "Prepare an external integration request."
	case StatusRequestReady:
		return "Emit the external integration request."
	case StatusAwaitingExternalSemanticIntegration:
		return "Give the integration request to an external agent."
	case StatusFailed:
		return "Review the failure and retry safely."
	default:
		return "Review workspace status before continuing."
	}
}
