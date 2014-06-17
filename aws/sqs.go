// Types and functions to push metrics to AWS Cloudwatch
package aws

import (
	"fmt"
	"github.com/soundcloud/sc-gaws/aws/credentials"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	sqsApiVersion = "2012-11-05"
)

// AwsStatsPusher implements the StatsPusher interface to allow
// pushing metrics to AWS Cloudwatch
type SqsClient struct {
	Endpoint string

	// AWS Credentials
	Credentials credentials.CredentialsProvider

	client *http.Client
}

func NewSqsClient(endpoint string, credentials credentials.CredentialsProvider) SqsClient {
	return SqsClient{endpoint, credentials, &http.Client{}}
}

func (s SqsClient) Publish(message string) error {

	urlStr := fmt.Sprintf("%s/?Action=SendMessage&Version=%s&Timestamp=%s&MessageBody=%s", s.Endpoint, sqsApiVersion, url.QueryEscape(timeInRfc3339(time.Now())), url.QueryEscape(message))
	req, _ := http.NewRequest("GET", urlStr, nil)
	sign(req, s.Credentials.GetCredentials())
	res, err := s.client.Do(req)
	if err != nil {
		return err
	} else {
		defer res.Body.Close()
		if res.StatusCode != 200 {
			body, _ := ioutil.ReadAll(res.Body)
			return fmt.Errorf("Publishing message failed with status code %d: %s", res.StatusCode, body)
		}
	}
	return nil
}
