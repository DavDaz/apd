package wiki

import "testing"

func TestNextActionForPendingLifecycle(t *testing.T) {
	cases := []struct {
		status Status
		want   string
	}{
		{StatusRegistered, "Prepare an external integration request."},
		{StatusRequestReady, "Emit the external integration request."},
		{StatusAwaitingExternalSemanticIntegration, "Give the integration request to an external agent."},
		{StatusFailed, "Review the failure and retry safely."},
	}
	for _, tc := range cases {
		t.Run(string(tc.status), func(t *testing.T) {
			if got := NextAction(tc.status); got != tc.want {
				t.Fatalf("NextAction(%q) = %q, want %q", tc.status, got, tc.want)
			}
		})
	}
}
