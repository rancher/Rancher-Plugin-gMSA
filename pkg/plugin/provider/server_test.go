package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aiyengar2/Rancher-Plugin-gMSA/pkg/plugin/provider/generated/norman/core/v1/fakes"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var mockServ *HTTPServer

func TestMain(m *testing.M) {
	mockServ = &HTTPServer{
		Credentials: &CredentialClient{
			Secrets: createMockClient(),
		},
	}
	mockServ.Engine = NewGinServer(mockServ, true)
	m.Run()
}

func createMockClient() *fakes.SecretInterfaceMock {
	return &fakes.SecretInterfaceMock{
		GetFunc: func(name string, opts metav1.GetOptions) (*v1.Secret, error) {
			switch name {
			case "test":
				return &v1.Secret{
					Data: map[string][]byte{
						"username":   []byte("one"),
						"password":   []byte("pass"),
						"domainName": []byte("test.com"),
					},
				}, nil
			case "unauthorizedTest":
				return nil, errors.NewForbidden(v1.Resource("secret"), "", nil)
			}
			return nil, errors.NewNotFound(v1.Resource("secret"), name)
		},
	}
}

func Test_Headers(t *testing.T) {
	type test struct {
		name         string
		expectedCode int
		headers      map[string]string
	}

	tests := []test{
		{
			name:         "empty headers",
			expectedCode: http.StatusBadRequest,
			headers:      nil,
		},
		{
			name:         "bad object header",
			expectedCode: http.StatusNotFound,
			headers: map[string]string{
				"object": "fake",
			},
		},
		{
			name:         "unauthorized test",
			expectedCode: http.StatusNotFound,
			headers: map[string]string{
				"object": "unauthorizedTest",
			},
		},
		{
			name:         "valid request",
			expectedCode: http.StatusOK,
			headers: map[string]string{
				"object": "test",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/provider", nil)
			if tc.headers != nil {
				for k, v := range tc.headers {
					req.Header.Set(k, v)
				}
			}
			mockServ.Engine.ServeHTTP(recorder, req)
			assert.Equal(t, recorder.Code, tc.expectedCode, fmt.Sprintf("case %s: expected response code to be %d but got %d", tc.name, tc.expectedCode, recorder.Code))
		})
	}
}

func Test_ResponseBody(t *testing.T) {
	type test struct {
		name               string
		headerObject       string
		expectedUsername   string
		expectedPassword   string
		expectedDomainName string
		secretName         string
	}

	tests := []test{
		{
			name:               "valid test",
			headerObject:       "test",
			expectedUsername:   "one",
			expectedPassword:   "pass",
			expectedDomainName: "test.com",
		},
		{
			name:         "not found test",
			headerObject: "wrong",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/provider", nil)
			req.Header.Set("object", tc.headerObject)
			mockServ.Engine.ServeHTTP(recorder, req)

			// if we got nothing but expected something
			if recorder.Body.Len() == 0 && (tc.expectedUsername != "" || tc.expectedPassword != "" || tc.expectedDomainName != "") {
				t.Error("received empty response body when a value was expected")
				return
			}
			// if we got nothing and expected nothing
			if recorder.Body.Len() == 0 && tc.expectedUsername == "" && tc.expectedPassword == "" && tc.expectedDomainName == "" {
				return
			}

			resp := Response{}
			assert.Nil(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
			assert.Equal(t, tc.expectedUsername, resp.Username)
			assert.Equal(t, tc.expectedPassword, resp.Password)
			assert.Equal(t, tc.expectedDomainName, resp.DomainName)
		})
	}
}
