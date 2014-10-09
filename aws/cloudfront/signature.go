package cloudfront

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// NewRSAPRivateKeyFromBytes get the first block of a PEM encoded bytes
// and return a rsa.PrivateKey
func NewRSAPrivateKeyFromBytes(b []byte) (*rsa.PrivateKey, error) {
	// ignore blocks other than the first one  as our file contains only one private key
	block, _ := pem.Decode(b)

	if block == nil {
		return nil, fmt.Errorf("Cannot decode PEM data")
	}

	// get a RSA private key from its ASN.1 PKCS#1 DER encoded form.
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse PKCS1 Private Key %v: %v", err)
	} else {
		return key, nil
	}
}

// NewRSAPrivateKeyFromFile call NewRSAPRivateKeyFromBytes on the content
// of filename, returning an rsa.PrivateKey
func NewRSAPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Cannot open private key file %v: %v", filename, err)
	}

	// decode a PEM file
	b, err := ioutil.ReadAll(f)

	if err != nil {
		return nil, fmt.Errorf("Cannot read file %v: %v", filename, err)
	}

	return NewRSAPrivateKeyFromBytes(b)
}

// SignPolicy return the proper signature needed to generate a valid Cloudfront
// Signed URL.
//
// More information:
//   http://goo.gl/pvA97e
// Command line equivalent:
//   cat policy | openssl sha1 -sign cloudfront-keypair.pem | openssl base64 | tr '+=/' '-_~'
func SignPolicy(privateKey *rsa.PrivateKey, policy CannedPolicy) (string, error) {
	signedPolicy, err := policy.signWithPrivateKey(privateKey)

	if err != nil {
		return "", fmt.Errorf("Cannot sign policy: %v", err)
	}

	signedPolicyBase64 := base64.StdEncoding.EncodeToString(signedPolicy)

	cloudfrontFmt := func(r rune) rune {
		switch r {
		case '+':
			return '-'
		case '=':
			return '_'
		case '/':
			return '~'
		case ' ':
			return 0
		default:
			return r
		}
	}

	return strings.Map(cloudfrontFmt, signedPolicyBase64), nil
}
