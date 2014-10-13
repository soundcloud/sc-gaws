package cloudfront

import "testing"

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

var expectedSignature = `ad9AA87ApevgO6cyNqlDqMOk0f35QCe7YbWyeY13vifI7xPEqyc85Nf9XBNtzyHZDD-u-nN5wcgByuMlMwN9yLDEV0B6AHVzKmcOyFuJKVqxGffrOh7LxxjpNImTJ1EgXHO-kY-buFHv5yoWraqGI7rMQ88a~dxViLrK7JL3yAQAV7HNes8dN8J1PIgTg1gSBKvNuH1QGjjyBJMcReY0kMJZsJVikHAx3N76xRzIPRBXd1-d9Q~6X6kWy8N2GuV9i4fpoa4Svwirz~HlvWjStuXN6Go7HSOnqLBCSLHnmwxhFrGFrqwTzNy671sw8VeT4NBZCrwrOtrfdAEevsen0w__`

func TestSignPolicy(t *testing.T) {
	pk, err := NewRSAPrivateKeyFromBytes([]byte(privateKey))

	if err != nil {
		t.Fatalf("%s", err)
	}

	signature, err := SignPolicy(pk, cannedPolicy)

	if err != nil {
		t.Fatalf("%v", err)
	}

	if signature != expectedSignature {
		t.Fatalf("Signature does not match\n--- Got ---\n%v\n--- Expected ---\n%v\n", signature, expectedSignature)
	}
}
