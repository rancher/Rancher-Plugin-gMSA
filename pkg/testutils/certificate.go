package testutils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	mrand "math/rand"
)

var (
	dummyOrg = []string{"Rancher CCG Plugin GMSA"}
)

func SelfSignedCertificate() (privateKeyPem []byte, publicCertificatePem []byte, err error) {
	privateKeyPemBuf, publicCertificatePemBuf := new(bytes.Buffer), new(bytes.Buffer)

	privateKeyRaw, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	privateKey := x509.MarshalPKCS1PrivateKey(privateKeyRaw)
	pem.Encode(privateKeyPemBuf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKey,
	})

	cert := &x509.Certificate{
		Subject: pkix.Name{
			Organization: dummyOrg,
		},
		SerialNumber: big.NewInt(mrand.Int63()),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	publicCertificate, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKeyRaw.PublicKey, privateKeyRaw)
	if err != nil {
		return nil, nil, err
	}
	pem.Encode(publicCertificatePemBuf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: publicCertificate,
	})
	return privateKeyPemBuf.Bytes(), publicCertificatePemBuf.Bytes(), err
}
