package server

import (
	"fmt"

	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
)

type Response struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DomainName string `json:"domainName"`
}

func ParseResponse(secret *corev1.Secret) (*Response, error) {
	if secret == nil {
		return nil, nil
	}
	var err error
	username, ok := secret.Data["username"]
	if !ok {
		err = multierr.Append(err, fmt.Errorf("does not contain key 'username'"))
	}
	password, ok := secret.Data["password"]
	if !ok {
		err = multierr.Append(err, fmt.Errorf("does not contain key 'password'"))
	}
	domainName, ok := secret.Data["domainName"]
	if !ok {
		err = multierr.Append(err, fmt.Errorf("does not contain key 'domainName'"))
	}
	if err == nil {
		return &Response{
			Username:   string(username),
			Password:   string(password),
			DomainName: string(domainName),
		}, nil
	}
	return nil, fmt.Errorf("error parsing response from secret %s: %s", secret.Name, err)
}
