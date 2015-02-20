package cloudfront

import (
	"sync"
	"testing"
	"time"
)

var (
	testURL        = "http://d5helwpxwoaq8.cloudfront.net/3tGKc9yN1IfN.128.mp3"
	testExpiryTime = time.Unix(1412799136, 0)
	testStartTime  = time.Unix(1412799000, 0)
	testSourceIP   = "192.0.2.1"

	// Signature, for convenience is base64 encoded
	// Generated with:
	// echo -n canned_policy_content | openssl sha1 -sign test_private_key.pem | openssl base64
	// where canned_policy_content is equal to expectedCannedPolicyString
	expectedCannedSignature = `Expires=1412799136&Signature=ad9AA87ApevgO6cyNqlDqMOk0f35QCe7YbWyeY13vifI7xPEqyc85Nf9XBNtzyHZDD-u-nN5wcgByuMlMwN9yLDEV0B6AHVzKmcOyFuJKVqxGffrOh7LxxjpNImTJ1EgXHO-kY-buFHv5yoWraqGI7rMQ88a~dxViLrK7JL3yAQAV7HNes8dN8J1PIgTg1gSBKvNuH1QGjjyBJMcReY0kMJZsJVikHAx3N76xRzIPRBXd1-d9Q~6X6kWy8N2GuV9i4fpoa4Svwirz~HlvWjStuXN6Go7HSOnqLBCSLHnmwxhFrGFrqwTzNy671sw8VeT4NBZCrwrOtrfdAEevsen0w__&Key-Pair-Id=APKA9ONS7QCOWEXAMPLE`
	expectedCustomSignature = `Policy=eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cDovL2Q1aGVsd3B4d29hcTguY2xvdWRmcm9udC5uZXQvM3RHS2M5eU4xSWZOLjEyOC5tcDMiLCJDb25kaXRpb24iOnsiRGF0ZUxlc3NUaGFuIjp7IkFXUzpFcG9jaFRpbWUiOjE0MTI3OTkxMzZ9LCJEYXRlR3JlYXRlclRoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTQxMjc5OTAwMH0sIklwQWRkcmVzcyI6eyJBV1M6U291cmNlSXAiOiIxOTIuMC4yLjEifX19XX0_&Signature=xEbhq3QI57fcobwlQL85esB2l-QuOroXdi15XTSXij2rw1LTxzjFIGK5w6f-KBscbO0vrVh~5DAbae7wPqRak0W3A-2h6IUV1AysggyIt0E1BCHjPe2iVDVdZdZ3tytEY1dAjv7kxZvqETzue8w3NyZqaW9lHIPKkMmn-Kyv0-2J5lFUymqud~LE~9ks65PReBwf-HPusPucb61cZdDWH7ccB4z4uGy5Osd09zR-yY9Q5cy65NkNZyd-z4UT-Aoj3CcIL9H4Zg6G7n1SufWaki4rNGdod-gTUhYeTqKbkI~SmzTphC6okRSBU~RvdSYde1GrpYts61KP6Vcts3GXRQ__&Key-Pair-Id=APKA9ONS7QCOWEXAMPLE`
)

// An RSA Private Key
// Generated with
//   ssh-keygen -t rsa -C "foo@bar.com""
var privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA5SuaCXhWtPuG6aGhLq8WbqM0LmFcXyqLfBQSlAQhxjIhV2dd
XJdDtdsZYUkOrp/H3M5FPAPH3zxYU+nRF3ek4rqmBjpGcQZuftY3nmxwTypYirQo
Iq73T3pr21DidIKV6nObcUfvLBawLiPd+TAg38j/VymuEUJhgfUC6WGIulb4ORfw
A+LymEz8LBF6bJpQ1owUCfhwujR6floyCxTp7a2DHwR7jcoLfexeFIzhHEP8UWd5
7J2tD5bgMqUFLQ34MtOGGD0QEAasw2i3lxOEcyUtTiYwCzV/MzpHD7G4m/zevqLV
IItBiSab8f2dAxyc1ubZGjNrFRhdnOYSJ6uX1QIDAQABAoIBADBZsacb15dZlg5G
xp313NK85i5+5iFB9anZBk5qTMHnI7ewHDeDxopgzosDAfD/zwgcEOlnlszXi38w
zqeX25bmcE7SDrib9cYW5icrk8pwEbw55Fnk9lKzbnwYJZ8VShHsEDinR6PSqZsi
gBup9tWgL5cxOQN1MONdUR7yMAm4F2KH3IWa2InNKaaBwvq9d/sAExlirLp7jlkR
R7jeEPxChodflBmzzb2NcEP+A8lrTLdrsfLwgLVPKNiwUoWym6lsKsGGmAtm5ZrA
94BbzAJNbQ3IwNpbayfplaHdYi6fBL4re4JWvBvQj0dKAZuC0ZyJ38BdpKhiiI1k
baGx3EECgYEA+z+JLP0U7hWKx9W29/WP3LkVK1GayLfMqi/XesvG143XsjHkMamd
D+eP+gFl8q5DF/znPpAGwwQHSswtsd4ZQAMGzl3r7TbC6fK1SVzAjmY1eOCKLBC/
DJW3a5EeRPmQkJ3zc+QXAKEp0MfbEO0M3bBXkmBP8YJWCO1k2Pb2sTECgYEA6YEr
/Gmz0SO6rd9j5+8GZ3FHUzBFGpx6WY/H9sSOyF2o2SrHBX5ZBUOKz2U4vxSAtqj0
2Q5zMBILHAWh1K/4MYx3ydJ4zVlLJe4JKqAXvrV3PiwSWTLVlZCEz1J0pIf7C5VE
R4XXUPwxZD7Kt2HL44ZmdwapMNkJX0I7iMzix+UCgYArHhE9jkU8QqgpeUzIKvVA
bObsIzoL/jb6cfFp2nTKY0ZEB3ng5/nTU+sKfZjwV+WdxUIuI2t1pkhWFso0vyfY
K2zMl6O4dvBmU8e2ylslVPcSQn6T51/SGhN7O1FVhvq/RswT9G3aJs5VTScUNYpC
tVOiBDNUAAkhyRPIhEF70QKBgFcBzsYLDPnM+m53Yt61bgl1aEJTJiy4Sc69iKEm
tJ1saNIi4m3lmn4foMWOzgpFNYDajAGlJL6wunsCjj6WI0EzOh910ZWSxDGp9Lhx
Vue07m24Nk4OZw+H/jTSYKw+DwlN4VdVQ9nlXvIfg6G5SyPInL2VOc390UQxhcV/
srERAoGAHjonqtL/OciIK9vkRf9IAjlthXumaEx6WofYp6avv8pir3VstV6KkP7o
kPvsrOhaj289FDUNfIEcziEPpWbMeMmq8DcYSkUsPjKYf5gqGK9qsbuhuFJEP48q
S7/p2V8ICzkHJ/8fOIKUC+Lu8VuiWueJXDmD/vMaghFkpC1yI4A=
-----END RSA PRIVATE KEY-----`

func TestSignCannedPolicy(t *testing.T) {
	pk, err := NewRSAPrivateKeyFromBytes([]byte(privateKey))
	if err != nil {
		t.Fatalf("%s", err)
	}

	cannedPolicy := NewCannedPolicy(testURL, testExpiryTime)
	signature, err := SignPolicy(pk, cannedPolicy, "APKA9ONS7QCOWEXAMPLE")
	if err != nil {
		t.Fatalf("%v", err)
	}

	if signature != expectedCannedSignature {
		t.Fatalf("signed policy does not match\n--- Got ---\n%v\n--- Expected ---\n%v\n", signature, expectedCannedSignature)
	}
}

func TestSignCustomPolicy(t *testing.T) {
	pk, err := NewRSAPrivateKeyFromBytes([]byte(privateKey))
	if err != nil {
		t.Fatalf("%s", err)
	}

	customPolicy := NewCustomPolicy(testURL, testExpiryTime).AddStartTime(testStartTime).AddSourceIP(testSourceIP)
	signature, err := SignPolicy(pk, customPolicy, "APKA9ONS7QCOWEXAMPLE")
	if err != nil {
		t.Fatalf("%v", err)
	}

	if signature != expectedCustomSignature {
		t.Fatalf("signed policy does not match\n--- Got ---\n%v\n--- Expected ---\n%v\n", signature, expectedCustomSignature)
	}
}

func BenchmarkSignCustomPolicy(b *testing.B) {
	pk, err := NewRSAPrivateKeyFromBytes([]byte(privateKey))
	if err != nil {
		b.Fatalf("%s", err)
	}

	for i := 0; i < b.N; i++ {
		customPolicy := NewCustomPolicy(testURL, testExpiryTime).AddStartTime(testStartTime).AddSourceIP(testSourceIP)
		_, err := SignPolicy(pk, customPolicy, "APKA9ONS7QCOWEXAMPLE")

		if err != nil {
			b.Fatalf("%v", err)
		}
	}
}

// The specific case we want to test here is when SignPKCS1v15
// is invoked by multiple go routines, which may ends up in a race condition
// if not properly handled.
func TestSSLSignRaceCondition(t *testing.T) {
	pk, err := NewRSAPrivateKeyFromBytes([]byte(privateKey))
	if err != nil {
		t.Fatalf("%s", err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 2000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			customPolicy := NewCustomPolicy(testURL, testExpiryTime).AddStartTime(testStartTime).AddSourceIP(testSourceIP)
			signature, err := SignPolicy(pk, customPolicy, "APKA9ONS7QCOWEXAMPLE")

			if err != nil {
				t.Fatalf("%v", err)
			}

			if signature != expectedCustomSignature {
				t.Fatalf("signed policy does not match\n--- Got ---\n%v\n--- Expected ---\n%v\n", signature, expectedCustomSignature)
			}
		}()
	}

	wg.Wait()
}
