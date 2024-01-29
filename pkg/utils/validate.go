package utils

import (
	"fmt"
	"regexp"
)

var (
	IsValidNamespace = regexp.MustCompile(`^[a-z]([a-z\-0-9]{1,61}[a-z0-9])?$`).MatchString
)

func ValidateNamespace(namespace string) error {
	if !IsValidNamespace(namespace) {
		return fmt.Errorf("gmsa-account-provider must be run within a kubernetes namespace")
	}
	return nil
}
