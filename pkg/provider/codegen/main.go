package main

import (
	"os"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/provider/codegen/generator"
	v1 "k8s.io/api/core/v1"
)

func main() {
	os.Unsetenv("GOPATH")

	generator.GenerateNativeTypes(v1.SchemeGroupVersion, []interface{}{
		v1.Secret{},
	}, nil)
}
