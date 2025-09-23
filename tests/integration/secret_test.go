package integrationpackage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider"
	"github.com/rancher/Rancher-Plugin-gMSA/pkg/provider/server"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/wrangler/v3/pkg/generated/controllers/core"
	v1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type ProviderSecretControllerTestSuite struct {
	suite.Suite
	ctx     context.Context
	cancel  context.CancelFunc
	testEnv *envtest.Environment

	providerCtx       context.Context
	providerCtxCancel context.CancelFunc

	secrets    v1.SecretController
	namespaces v1.NamespaceController

	client *restclient.Config
}

func TestSecretController(t *testing.T) {
	suite.Run(t, new(ProviderSecretControllerTestSuite))
}

func (p *ProviderSecretControllerTestSuite) SetupSuite() {
	p.ctx, p.cancel = context.WithCancel(context.Background())

	p.testEnv = &envtest.Environment{}
	restCfg, err := p.testEnv.Start()
	assert.NoError(p.T(), err)
	assert.NotNil(p.T(), restCfg)
	p.client = restCfg

	controllerFactory, err := controller.NewSharedControllerFactoryFromConfigWithOptions(p.client, runtime.NewScheme(), nil)
	if err != nil {
		p.T().Fatal(err)
	}

	opts := &generic.FactoryOptions{
		SharedControllerFactory: controllerFactory,
	}

	coreFactory, err := core.NewFactoryFromConfigWithOptions(p.client, opts)
	if err != nil {
		p.T().Fatal(err)
	}

	p.secrets = coreFactory.Core().V1().Secret()
	p.namespaces = coreFactory.Core().V1().Namespace()
}

func (p *ProviderSecretControllerTestSuite) TearDownSuite() {
	err := p.testEnv.Stop()
	require.NoError(p.T(), err)
}

func (p *ProviderSecretControllerTestSuite) TestSecretRetrieval() {
	// create a namespace
	// create a secret in that namespace with a good format
	// create one with a bad format
	testNS := "secret-ns"
	p.createNS(testNS)

	provCtx, provCtxCancel := context.WithCancel(p.ctx)
	defer provCtxCancel()

	port := "3344"
	var err error
	go func() {
		err = provider.Run(provCtx, p.client, provider.Opts{
			Namespace:     testNS,
			ForcedPort:    port,
			DisableMTLS:   true,
			SkipArtifacts: true,
		})
		if err != nil {
			p.T().Fatal(err)
		}
		p.T().Log(fmt.Sprintf("Server started on port %s", port))
	}()

	time.Sleep(time.Second * 5)

	tests := []struct {
		name                 string
		data                 map[string]string
		expectedResponseCode int
	}{
		{
			name: "valid-account",
			data: map[string]string{
				"username":   "something",
				"password":   "password123",
				"domainName": "rancher.com",
			},
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "no-username-account",
			data: map[string]string{
				"password":   "abc123",
				"domainName": "Rancher",
			},
			expectedResponseCode: http.StatusNotFound,
		},
		{
			name: "no-password-account",
			data: map[string]string{
				"username":   "testuser",
				"domainName": "Rancher",
			},
			expectedResponseCode: http.StatusNotFound,
		},
		{
			name: "no-domain-name-account",
			data: map[string]string{
				"username": "testuser",
				"password": "Rancher",
			},
			expectedResponseCode: http.StatusNotFound,
		},
	}

	for _, t := range tests {
		tt := t
		p.T().Run(t.name, func(t *testing.T) {
			p.createSecret(tt.name, testNS, tt.data)
			resp := p.requestAndCheckStatus(port, tt.name, tt.expectedResponseCode)
			if resp != nil {
				assert.Equal(t, tt.data["username"], resp.Username)
				assert.Equal(t, tt.data["password"], resp.Password)
				assert.Equal(t, tt.data["domainName"], resp.DomainName)
			}
		})
	}
}

func (p *ProviderSecretControllerTestSuite) requestAndCheckStatus(port, secretName string, expectedStatus int) *server.Response {
	url := fmt.Sprintf("http://127.0.0.1:%s/provider", port)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		p.T().Fatal(err)
	}
	req.Header.Set("object", secretName)
	p.T().Log(fmt.Sprintf("making request to %s", url))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		p.T().Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		p.T().Fail()
		return nil
	}

	if expectedStatus != http.StatusOK {
		return nil
	}

	serverResp := &server.Response{}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		p.T().Fail()
		return nil
	}

	err = json.Unmarshal(b, serverResp)
	if err != nil {
		p.T().Fail()
		return nil
	}

	return serverResp
}

func (p *ProviderSecretControllerTestSuite) createNS(name string) {
	p.T().Log(fmt.Sprintf("creating namespace %s", name))
	_, err := p.namespaces.Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	if err != nil {
		p.T().Fatal(err)
	}
}

func (p *ProviderSecretControllerTestSuite) createSecret(name, namespace string, contents map[string]string) {
	p.T().Log(fmt.Sprintf("creating secret %s in namespace %s with content %v", name, namespace, contents))
	_, err := p.secrets.Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: contents,
	})
	if err != nil {
		p.T().Fatal(err)
	}
}
