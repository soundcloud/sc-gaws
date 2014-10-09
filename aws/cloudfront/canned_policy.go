package cloudfront

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"time"
)

const (
	cannedPolicyFmt = `{"Statement":[{"Resource":"%v","Condition":{"DateLessThan":{"AWS:EpochTime":%v}}}]}`
)

type CannedPolicy struct {
	Url       string
	ExpiresAt time.Time
}

func (p CannedPolicy) String() string {
	return fmt.Sprintf(cannedPolicyFmt, p.Url, p.ExpiresAt.Unix())
}

func (p CannedPolicy) signWithPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	// create a sha1 digest of our policy
	hashFunc := crypto.SHA1
	h := hashFunc.New()
	io.WriteString(h, p.String())
	digest := h.Sum(nil)

	// calculates the signature of digest using RSASSA-PKCS1-V1_5-SIGN from RSA PKCS#1 v1.5.
	if signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hashFunc, digest); err != nil {
		return []byte{}, err
	} else {
		return signature, nil
	}
}
