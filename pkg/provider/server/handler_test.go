package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultNamespace = "cattle-windows-gmsa-system"
	testCases        = map[string]testCase{
		toKey(defaultNamespace, "test-secret"): {
			Secret: secretInput{
				Namespace:     defaultNamespace,
				Name:          "test-secret",
				HasUsername:   true,
				HasPassword:   true,
				HasDomainName: true,
			}.Secret(),
			ExpectedStatusCode: http.StatusOK,
		},
		toKey(defaultNamespace, "no-username"): {
			Secret: secretInput{
				Namespace:     defaultNamespace,
				Name:          "no-username",
				HasUsername:   false,
				HasPassword:   true,
				HasDomainName: true,
			}.Secret(),
			ExpectedStatusCode: http.StatusNotFound,
		},
		toKey(defaultNamespace, "no-password"): {
			Secret: secretInput{
				Namespace:     defaultNamespace,
				Name:          "no-password",
				HasUsername:   true,
				HasPassword:   false,
				HasDomainName: true,
			}.Secret(),
			ExpectedStatusCode: http.StatusNotFound,
		},
		toKey(defaultNamespace, "no-domain-name"): {
			Secret: secretInput{
				Namespace:     defaultNamespace,
				Name:          "no-domain-name",
				HasUsername:   true,
				HasPassword:   true,
				HasDomainName: false,
			}.Secret(),
			ExpectedStatusCode: http.StatusNotFound,
		},
		toKey(defaultNamespace, "no-content"): {
			Secret: secretInput{
				Namespace:     defaultNamespace,
				Name:          "no-content",
				HasUsername:   false,
				HasPassword:   false,
				HasDomainName: false,
			}.Secret(),
			ExpectedStatusCode: http.StatusNotFound,
		},
		toKey("different-namespace", "test-secret-2"): {
			Secret: secretInput{
				Namespace:     "different-namespace",
				Name:          "test-secret-2",
				HasUsername:   false,
				HasPassword:   false,
				HasDomainName: false,
			}.Secret(),
			// Should not be able to find secret outside of configured namespace
			ExpectedStatusCode: http.StatusNotFound,
		},
	}
)

type testCase struct {
	Secret             *v1.Secret
	ExpectedStatusCode int
}

func TestHandler(t *testing.T) {
	g := &secretsGetter{}
	handler := NewHandler(g, defaultNamespace)
	keyTestFunc := func(key string, tc testCase, method, url, headerKey string) func(*testing.T) {
		return func(t *testing.T) {
			// parse secret into expected response
			expectedResponse, responseErr := ParseResponse(tc.Secret)
			var expectedResponseBody []byte
			if expectedResponse != nil && responseErr == nil {
				var err error
				expectedResponseBody, err = json.Marshal(expectedResponse)
				assert.Nil(t, err, "failed to marshal expected response")
			}

			// set up mock request and response
			name := strings.Split(key, "/")[1]
			r, err := http.NewRequest(method, url, nil)
			assert.Nil(t, err, "failed to create HTTP request")
			if headerKey != "" {
				r.Header.Set(headerKey, name)
			}
			w := httptest.NewRecorder()

			// Run handler
			handler.ServeHTTP(w, r)
			body := w.Body.String()

			assert.Equal(t, tc.ExpectedStatusCode, w.Code)
			if expectedResponse != nil && responseErr == nil {
				assert.Equal(t, string(expectedResponseBody), body)
			}
		}
	}
	for key, tc := range testCases {
		t.Run(key, keyTestFunc(key, tc, "GET", "/provider", "object"))
	}
	key := toKey(defaultNamespace, "nonexistent")
	t.Run("Nonexistent", keyTestFunc(key, testCase{
		ExpectedStatusCode: http.StatusNotFound,
	}, "GET", "/provider", "object"))

	key = toKey(defaultNamespace, "test-secret")
	t.Run("POST Request", keyTestFunc(key, testCase{
		ExpectedStatusCode: http.StatusNotFound,
	}, "POST", "/provider", "object"))

	t.Run("Invalid Endpoint", keyTestFunc(key, testCase{
		ExpectedStatusCode: http.StatusNotFound,
	}, "GET", "/invalid", "object"))

	t.Run("Invalid header", keyTestFunc(key, testCase{
		ExpectedStatusCode: http.StatusNotFound,
	}, "GET", "/provider", "object2"))

	t.Run("Missing header", keyTestFunc(key, testCase{
		ExpectedStatusCode: http.StatusNotFound,
	}, "GET", "/provider", ""))
}

type secretInput struct {
	Namespace     string
	Name          string
	HasUsername   bool
	HasPassword   bool
	HasDomainName bool
}

func (i secretInput) Secret() *corev1.Secret {
	if i.Name == "" || i.Namespace == "" {
		logrus.Fatalf("invalid secretInput: %v", i)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      i.Name,
		},
		Data: map[string][]byte{},
	}
	if i.HasUsername {
		secret.Data["username"] = []byte("username")
	}
	if i.HasPassword {
		secret.Data["password"] = []byte("password")
	}
	if i.HasDomainName {
		secret.Data["domainName"] = []byte("ad.domain")
	}
	return secret
}

func toKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

type secretsGetter struct {
}

func (s *secretsGetter) Get(namespace, name string) (*corev1.Secret, error) {
	key := toKey(namespace, name)
	secret, ok := testCases[key]
	if !ok {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    "",
			Resource: "secrets",
		}, name)
	}
	return secret.Secret, nil
}
