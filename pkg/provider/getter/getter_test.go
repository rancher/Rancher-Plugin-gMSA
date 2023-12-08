package getter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

type testGetter struct {
}

func (t *testGetter) Get(namespace, name string) (runtime.Object, error) {
	return nil, fmt.Errorf("%s/%s", namespace, name)
}

func TestNamespaced(t *testing.T) {
	g := &testGetter{}
	namespaced := Namespaced[runtime.Object](g, "hello")

	_, err := namespaced.Get("world")
	assert.Equal(t, "hello/world", err.Error())

	_, err = namespaced.Get("rancher")
	assert.Equal(t, "hello/rancher", err.Error())
}
