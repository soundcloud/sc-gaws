package cloudfront

import (
	"fmt"
	"strings"
	"time"
)

const (
	cannedPolicyFmt = `{"Statement":[{"Resource":"%v","Condition":{"DateLessThan":{"AWS:EpochTime":%v}}}]}`
)

type CannedPolicy struct {
	Url       string
	ExpiresAt time.Time
}

// NewCannedPolicy returns a new CannedPolicy with the given URL and Expiry
// time.
func NewCannedPolicy(url string, expiry time.Time) CannedPolicy {
	// Throw away any query string parameters
	urlParts := strings.SplitN(url, "?", 2)

	return CannedPolicy{urlParts[0], expiry}
}

// String returns the compacted-JSON format of the Canned Policy.
func (p CannedPolicy) String() string {
	return fmt.Sprintf(cannedPolicyFmt, p.Url, p.ExpiresAt.Unix())
}

// signWithPrivateKey returns a binary encoding of the Canned Policy signature
func (p CannedPolicy) signWithPrivateKey(privateKey PrivateKey) ([]byte, error) {
	// calculates the signature of digest using RSASSA-PKCS1-V1_5-SIGN from RSA PKCS#1 v1.5.
	return privateKey.SignPKCS1v15([]byte(p.String()))
}
