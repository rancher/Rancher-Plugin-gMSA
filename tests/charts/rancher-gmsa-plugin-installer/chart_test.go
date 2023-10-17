package simple

import (
	"testing"

	"github.com/rancher/hull/pkg/test"
)

func TestChart(t *testing.T) {
	opts := test.GetRancherOptions()
	opts.YAMLLint.Enabled = true
	opts.HelmLint.Rancher.Enabled = false // TODO: remove this and fix here once https://github.com/rancher/hull/pull/41 is merged
	opts.Coverage.Disabled = true         // TODO: remove this when full tests are added
	suite.Run(t, opts)
}
