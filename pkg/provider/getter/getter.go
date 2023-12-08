package getter

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type NamespacedGeneric[T runtime.Object] interface {
	Get(name string) (T, error)
}

type Generic[T runtime.Object] interface {
	Get(namespace, name string) (T, error)
}

type NamespacedGenericFunc[T runtime.Object] func(name string) (T, error)

func (f NamespacedGenericFunc[T]) Get(name string) (T, error) {
	return f(name)
}

func Namespaced[T runtime.Object](getter Generic[T], namespace string) NamespacedGeneric[T] {
	getFunc := func(name string) (T, error) {
		return getter.Get(namespace, name)
	}
	return NamespacedGenericFunc[T](getFunc)
}
