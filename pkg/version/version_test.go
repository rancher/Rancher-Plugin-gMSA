package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FriendlyVersion returns the version to be displayed on running --version
func TestFriendlyVersion(t *testing.T) {
	testCases := []struct {
		Name      string
		Version   string
		GitCommit string

		Expected string
	}{
		{
			Name: "Default",
			// leaving this as a fmt.Sprintf since can be affected by build flags
			Expected: fmt.Sprintf("%s (%s)", Version, GitCommit),
		},
		{
			Name:      "GA",
			Version:   "v1.0.0",
			GitCommit: "<commit-hash>",
			Expected:  "v1.0.0 (<commit-hash>)",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			version, gitCommit := Version, GitCommit
			defer func() {
				Version = version
				GitCommit = gitCommit
			}()
			if tc.Version != "" {
				Version = tc.Version
			}
			if tc.GitCommit != "" {
				GitCommit = tc.GitCommit
			}
			assert.Equal(t, tc.Expected, FriendlyVersion())
		})
	}
}
