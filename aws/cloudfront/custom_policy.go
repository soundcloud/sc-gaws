package cloudfront

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"encoding/json"
	"io"
	"strings"
	"time"
)

// CustomPolicy represents a CloudFront Custom Policy.
// It supports resource URLs with wildcards, starting and ending validity times
// and source IP restrictions.
type CustomPolicy struct {
	Statement []customPolicyStatement
}

type customPolicyStatement struct {
	Resource  string
	Condition customPolicyCondition
}

type customPolicyCondition struct {
	DateLessThan    *customPolicyDate
	DateGreaterThan *customPolicyDate    `json:",omitempty"`
	IpAddress       *customPolicyAddress `json:",omitempty"`
}

type customPolicyDate struct {
	EpochTime int64 `json:"AWS:EpochTime"`
}

type customPolicyAddress struct {
	SourceIP string `json:"AWS:SourceIp"`
}

// NewCustomPolicy generates a new CustomPolicy based on the mandatory
// arguments for the resource URL and the expiry time.
func NewCustomPolicy(url string, expiresAt time.Time) CustomPolicy {
	// Throw away any query string parameters
	urlParts := strings.SplitN(url, "?", 2)

	return CustomPolicy{
		[]customPolicyStatement{
			customPolicyStatement{
				urlParts[0],
				customPolicyCondition{
					DateLessThan:    &customPolicyDate{expiresAt.Unix()},
					DateGreaterThan: nil,
					IpAddress:       nil,
				},
			},
		},
	}
}

// AddStartTime adds a DateGreaterThan restriction to the custom policy
// condition. It is suitable for chaining.
func (p CustomPolicy) AddStartTime(startsAt time.Time) CustomPolicy {
	p.Statement[0].Condition.DateGreaterThan = &customPolicyDate{startsAt.Unix()}
	return p
}

// AddSourceIP adds a Source IP Address restriction to the custom policy
// condition. It is suitable for chaining.
func (p CustomPolicy) AddSourceIP(sourceIP string) CustomPolicy {
	p.Statement[0].Condition.IpAddress = &customPolicyAddress{sourceIP}
	return p
}

// String returns the compacted-JSON format of the Custom Policy.
func (p CustomPolicy) String() string {
	encodedPolicy, err := json.Marshal(p)
	if err != nil {
		return ""
	}

	var compactedPolicy bytes.Buffer
	err = json.Compact(&compactedPolicy, encodedPolicy)
	if err != nil {
		return ""
	}

	return compactedPolicy.String()
}

// signWithPrivateKey returns a binary encoding of the Custom Policy signature
func (p CustomPolicy) signWithPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	// create a sha1 digest of our policy
	hashFunc := crypto.SHA1
	h := hashFunc.New()
	io.WriteString(h, p.String())
	digest := h.Sum(nil)

	// calculates the signature of digest using RSASSA-PKCS1-V1_5-SIGN from RSA PKCS#1 v1.5.
	if signature, err := rsa.SignPKCS1v15(nil, privateKey, hashFunc, digest); err != nil {
		return []byte{}, err
	} else {
		return signature, nil
	}
}
