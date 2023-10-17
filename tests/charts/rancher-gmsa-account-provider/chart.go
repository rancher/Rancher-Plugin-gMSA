package simple

import (
	"github.com/rancher/hull/pkg/chart"
	"github.com/rancher/hull/pkg/test"
	"github.com/rancher/hull/pkg/utils"
)

var ChartPath = utils.MustGetLatestChartVersionPathFromIndex("../index.yaml", "rancher-gmsa-account-provider", true)

var (
	DefaultReleaseName = "rancher-gmsa-account-provider"
	DefaultNamespace   = "cattle-windows-gmsa-system"
)

var suite = test.Suite{
	ChartPath: ChartPath,

	Cases: []test.Case{
		{
			Name: "Using Defaults",

			TemplateOptions: chart.NewTemplateOptions(DefaultReleaseName, DefaultNamespace),
		},
	},

	NamedChecks: []test.NamedCheck{},

	FailureCases: []test.FailureCase{},
}
