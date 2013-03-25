// Provides a common interface to deal with different types of AWS Credentials.
//
// This package supports two different of credentials
//
// * Regular IAM User credentials

// * Temporary credentials  which are associated with an IAM Role, fetched
//   from the EC2 Metadata Service.
//   More info about temporary credentials: http://docs.aws.amazon.com/STS/latest/UsingSTS/UsingTokens.html

package credentials

import (
	"sync"
)

// Credentials type to hold AWS Credentials
type Credentials struct {
	AccessKeyId     string
	SecretAccessKey string
	Token           string // Security token if applicable. Blank if not used
}

// CredentialsProvider is an interface that wraps a method Credentials which
// can be called to obtain AWS Credentials that can be used to authenticate
// against AWS Services
type CredentialsProvider interface {
	// FIXME possibly return an error if credentials cannot be retrieved
	// for some reason.
	GetCredentials() *Credentials
}

// Credentials obtained from the EC2 Metadata API contain some
// additional information.
type ec2MetadataCredentials struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
	mu              *sync.Mutex // Mutex to synchronize credential refresh
}

// Simply returns itself
func (c *Credentials) GetCredentials() *Credentials {
	return c
}

// Returns a copy of the credentials obtained from the EC2 Metadata
// API.
//
// Needs to synchronize access to the credentials between the
// current goroutine and the goroutine that is fetching the
// credentials periodically from the EC2 Metadata API.
func (c *ec2MetadataCredentials) GetCredentials() *Credentials {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &Credentials{AccessKeyId: c.AccessKeyId,
		SecretAccessKey: c.SecretAccessKey,
		Token:           c.Token,
	}
}

// Initialise regular IAM Credentials
func NewIamUserCredentials(keyId string, secretKey string) CredentialsProvider {
	return &Credentials{AccessKeyId: keyId, SecretAccessKey: secretKey}
}

// Initialise role credentials. Returns an error if credentials could not be
// initalized correctly.
func NewIamRoleCredentials(role string) (CredentialsProvider, error) {
	creds := &ec2MetadataCredentials{mu: &sync.Mutex{}}
	err := fetchRoleCredentials(role, creds)
	if err != nil {
		return nil, err
	}

	go credentialsRefresher(creds, role)
	return creds, nil
}
