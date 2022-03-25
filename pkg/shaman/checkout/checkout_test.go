package checkout

import (
	"testing"
)

func Test_isValidCheckoutPath(t *testing.T) {
	tests := []struct {
		name         string
		checkoutPath string
		want         bool
	}{
		// Valid cases.
		{"simple", "a", true},
		{"uuid", "5e5be786-e6d7-480c-90e6-437f9ef5bf5d", true},
		{"with-spaces", "5e5be786 e6d7 480c 90e6 437f9ef5bf5d", true},
		{"project-scene-job-discriminator", "Sprite-Fright/scenename/jobname/2022-03-25-11-30-feb3", true},
		{"unicode", "ránið/lélegt vélmenni", true},

		// Invalid cases.
		{"empty", "", false},
		{"backslashes", "with\\backslash", false},
		{"windows-drive-letter", "c:/blah", false},
		{"question-mark", "blah?", false},
		{"star", "blah*hi", false},
		{"semicolon", "blah;hi", false},
		{"colon", "blah:hi", false},
		{"absolute-path", "/blah", false},
		{"directory-up", "path/../../../../etc/passwd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidCheckoutPath(tt.checkoutPath); got != tt.want {
				t.Errorf("isValidCheckoutPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
