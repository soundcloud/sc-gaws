package aws

import (
	"github.com/soundcloud/sc-gaws/aws/credentials"
	"os"
	"testing"
)

func TestPublish(t *testing.T) {

	var (
		awsAkId     = os.Getenv("AWS_ACCESS_KEY_ID")
		awsSecretAk = os.Getenv("AWS_SECRET_ACCESS_KEY")
		endpoint    = os.Getenv("SQS_TEST_ENDPOINT")
	)

	if endpoint == "" {
		t.Fatalf("Endpoint URL must be set.")
	}

	sqlClient := NewSqsClient(endpoint, credentials.NewIamUserCredentials(awsAkId, awsSecretAk))
	err := sqlClient.Publish("Testdsfsdfdsäfdksfäösdf843009348093853089")

	if err != nil {
		t.Fatalf(err.Error())
	}
}
