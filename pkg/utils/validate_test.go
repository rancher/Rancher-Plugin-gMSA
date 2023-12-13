package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNamespace(t *testing.T) {
	testCases := []struct {
		Name      string
		Namespace string

		ExpectError bool
	}{
		{
			Name:      "Default",
			Namespace: "default",
		},
		{
			Name:      "One Character",
			Namespace: "a",
		},
		{
			Name:      "63 characters",
			Namespace: strings.Repeat("a", 63),
		},
		{
			Name:      "64 characters",
			Namespace: strings.Repeat("a", 64),

			ExpectError: true,
		},
		{
			Name:      "Numbers and Dashes",
			Namespace: "abcd-39920048-d-sdf--sdf-f--f",
		},
		{
			Name:        "Numbers and Dashes With Leading Dash",
			Namespace:   "-abcd-39920048-d-sdf--sdf-f--f",
			ExpectError: true,
		},
		{
			Name:        "Numbers and Dashes With Trailing Dash",
			Namespace:   "abcd-39920048-d-sdf--sdf-f--f-",
			ExpectError: true,
		},
		{
			Name:        "Numbers and Dashes With Leading Number",
			Namespace:   "3-abcd-39920048-d-sdf--sdf-f--f",
			ExpectError: true,
		},
		{
			Name: "Numbers and Dashes With Trailing Number",
			// this is valid
			Namespace: "abcd-39920048-d-sdf--sdf-f--f3",
		},
		{
			Name:        "Uppercase",
			Namespace:   "ABCDEEF",
			ExpectError: true,
		},
		{
			Name:        "Uppercase With Lowercase",
			Namespace:   "abCDrf",
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.ExpectError {
				assert.NotNil(t, ValidateNamespace(tc.Namespace), "expected this to be an invalid namespace")
			} else {
				assert.Nil(t, ValidateNamespace(tc.Namespace), "expected this to be an valid namespace")
			}
		})
	}
}
