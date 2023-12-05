package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	successfulResponse = &Response{
		Username:   "username",
		Password:   "password",
		DomainName: "ad.domain",
	}
)

func TestResponse(t *testing.T) {
	for key, tc := range testCases {
		t.Run(key, func(t *testing.T) {
			response, err := ParseResponse(tc.Secret)
			if tc.ExpectedStatusCode != http.StatusOK {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Equal(t, successfulResponse, response)
				assert.Nil(t, err, "did not expect error")
			}
		})
	}
	t.Run("Nil", func(t *testing.T) {
		response, err := ParseResponse(nil)
		assert.Nil(t, response)
		assert.Nil(t, err, "did not expect error")
	})
}
