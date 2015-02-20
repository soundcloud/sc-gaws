package cloudfront

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// CloudFront uses its own Base64 encoding
const encodeCloudFront = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-~"

// PolicySigner is the interface implemented by an object that can return a
// signed string representing its CloudFront policy.
type PolicySigner interface {
	signWithPrivateKey(PrivateKey) ([]byte, error)
	String() string
}

// NewRSAPRivateKeyFromBytes get the first block of a PEM encoded bytes
// and return a rsa.PrivateKey
func NewRSAPrivateKeyFromBytes(b []byte) (PrivateKey, error) {
	return LoadPrivateKeyFromPEM(b)
}

// NewRSAPrivateKeyFromFile call NewRSAPRivateKeyFromBytes on the content
// of filename, returning an rsa.PrivateKey
func NewRSAPrivateKeyFromFile(filename string) (PrivateKey, error) {
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

// SignPolicy return the proper signature and other parameters needed to
// generate a valid Cloudfront Signed URL.
// For canned policies this is: Expires, Signature, Key-Pair-Id
// For custom policies this is: Policy, Signature, Key-Pair-Id
//
// More information:
//   http://goo.gl/pvA97e
// Command line equivalent:
//   cat policy | openssl sha1 -sign cloudfront-keypair.pem | openssl base64 | tr '+=/' '-_~'
func SignPolicy(privateKey PrivateKey, policy PolicySigner, keyPairID string) (string, error) {
	signature, err := policy.signWithPrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("Cannot sign policy: %v", err)
	}

	encoding := base64.NewEncoding(encodeCloudFront)
	paddingMap := func(r rune) rune {
		switch r {
		case '=':
			return '_'
		default:
			return r
		}
	}

	switch policy.(type) {
	case CannedPolicy:
		return fmt.Sprintf("Expires=%d&Signature=%s&Key-Pair-Id=%s",
			policy.(CannedPolicy).ExpiresAt.Unix(),
			strings.Map(paddingMap, encoding.EncodeToString(signature)),
			keyPairID,
		), nil
	case CustomPolicy:
		return fmt.Sprintf("Policy=%s&Signature=%s&Key-Pair-Id=%s",
			strings.Map(paddingMap, encoding.EncodeToString([]byte(policy.String()))),
			strings.Map(paddingMap, encoding.EncodeToString(signature)),
			keyPairID,
		), nil
	}
	return "", nil
}
