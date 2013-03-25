// Implements a way to fetch temporary AWS credentials from the EC2 Metadata of
// a running instance that was launched with an IAM Role.
//
// A long running goroutine queries the EC2 Metadata via the web API, and
// extracts the credentials. The goroutine ensures that a refresh of
// credentials is initiated before the current credentails expire.
//
// More info: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/UsingIAM.html#UsingIAMrolesWithAmazonEC2Instances

package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	// Prefix for the URL to hit the EC2 Metadata API with
	credentialsEndpoint = "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
)

// Refresh IAM Role credentials for EC2 Metadata API before they expire
func credentialsRefresher(c *ec2MetadataCredentials, role string) {
	for {
		expiry, err := time.Parse(time.RFC3339, c.Expiration)
		if err != nil {
			log.Fatalf("Could not parse expiration - %s : %s", c.Expiration, err)
		}
		duration := expiry.Sub(time.Now())
		if duration < 0 {
			log.Printf("Credentials already expired at %s. Initiating refresh", c.Expiration)
			duration = 1
		}

		t := time.After(duration - (1 * time.Second))
		log.Printf("Refreshing credentials in %s", duration)
		select {
		case <-t:
			fetchRoleCredentials(role, c)
			break
		}
	}
}

// Fetch credentials for a role
func fetchRoleCredentials(role string, creds *ec2MetadataCredentials) error {

	url := credentialsEndpoint + role
	log.Printf("Querying EC2 Metadata for credentials: %s", url)
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("EC2 Metadata returned status %s\n%s", resp.StatusCode, body))
	}

	creds.mu.Lock()
	defer creds.mu.Unlock()
	err = json.Unmarshal(body, creds)
	if err != nil {
		return err
	}

	return nil
}