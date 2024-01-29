package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelfSignedCertificate(t *testing.T) {
	privateKey, publicCertificate, err := SelfSignedCertificate()
	assert.Nil(t, err, "encountered error while generating dummy certificates")
	assert.NotEmpty(t, string(privateKey), "found empty private key")
	assert.NotEmpty(t, string(publicCertificate), "found empty private key")
}
