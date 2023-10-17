package simple

import (
	"testing"

	"github.com/rancher/hull/pkg/test"
)

func TestChart(t *testing.T) {
	opts := test.GetRancherOptions()
	opts.YAMLLint.Enabled = true
	opts.Coverage.Disabled = true // TODO: remove this when full tests are added
	suite.Run(t, opts)
}
