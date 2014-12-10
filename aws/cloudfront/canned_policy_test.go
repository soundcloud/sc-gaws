package cloudfront

import (
	"encoding/base64"
	"testing"
)

var (
	// This is the Canned Policy we are going to test with
	expectedCannedPolicyString    = `{"Statement":[{"Resource":"http://d5helwpxwoaq8.cloudfront.net/3tGKc9yN1IfN.128.mp3","Condition":{"DateLessThan":{"AWS:EpochTime":1412799136}}}]}`
	expectedCannedPolicySignature = `ad9AA87ApevgO6cyNqlDqMOk0f35QCe7YbWyeY13vifI7xPEqyc85Nf9XBNtzyHZDD+u+nN5wcgByuMlMwN9yLDEV0B6AHVzKmcOyFuJKVqxGffrOh7LxxjpNImTJ1EgXHO+kY+buFHv5yoWraqGI7rMQ88a/dxViLrK7JL3yAQAV7HNes8dN8J1PIgTg1gSBKvNuH1QGjjyBJMcReY0kMJZsJVikHAx3N76xRzIPRBXd1+d9Q/6X6kWy8N2GuV9i4fpoa4Svwirz/HlvWjStuXN6Go7HSOnqLBCSLHnmwxhFrGFrqwTzNy671sw8VeT4NBZCrwrOtrfdAEevsen0w==`
)

func TestCannedPolicyStringer(t *testing.T) {
	cannedPolicy := NewCannedPolicy(testURL, testExpiryTime)
	cannedPolicyString := cannedPolicy.String()

	if cannedPolicyString != expectedCannedPolicyString {
		t.Fatalf("Invalid canned policy")
	}
}

func TestCannedPolicySignWithPrivateKey(t *testing.T) {
	cannedPolicy := NewCannedPolicy(testURL, testExpiryTime)

	// Create an rsa.PrivateKey
	pk, err := NewRSAPrivateKeyFromBytes([]byte(privateKey))
	if err != nil {
		t.Fatalf("%s", err)
	}

	// Sign it
	signed, err := cannedPolicy.signWithPrivateKey(pk)
	if err != nil {
		t.Fatalf("%s", err)
	}

	// encode our signature in base64, to make it easier to compare with the
	// binary output.
	signedb64 := base64.StdEncoding.EncodeToString([]byte(signed))

	// compare it with the one generated by openssl
	if signedb64 != expectedCannedPolicySignature {
		t.Fatalf("signature does not match\n--- Got ---\n%v\n--- Expected ---\n%v\n", signedb64, expectedCannedPolicySignature)
	}
}
