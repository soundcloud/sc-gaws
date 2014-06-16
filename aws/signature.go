package aws

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/soundcloud/sc-gaws/aws/credentials"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var b64 = base64.StdEncoding

// Version 2 signing for AWS Requests (http://goo.gl/RSRp5)
func sign(req *http.Request, keys *credentials.Credentials) {

	params := req.URL.Query()

	params.Set("AWSAccessKeyId", keys.AccessKeyId)
	params.Set("SignatureVersion", "2")
	params.Set("SignatureMethod", "HmacSHA256")

	// Check if we are using temporary credentials
	if keys.Token != "" {
		params.Set("SecurityToken", keys.Token)
	}

	var sarray []string
	for k, _ := range params {
		sarray = append(sarray, url.QueryEscape(k)+"="+url.QueryEscape(params.Get(k)))
	}
	sort.StringSlice(sarray).Sort()
	joined := strings.Join(sarray, "&")
	payload := req.Method + "\n" + req.Host + "\n" + req.URL.Path + "\n" + joined
	// log.Print(payload)
	hash := hmac.New(sha256.New, []byte(keys.SecretAccessKey))
	hash.Write([]byte(payload))
	signature := make([]byte, b64.EncodedLen(hash.Size()))
	b64.Encode(signature, hash.Sum(nil))

	params.Set("Signature", string(signature))

	req.URL.RawQuery = params.Encode()
}

// Convert time to RFC 3339 format
func timeInRfc3339(t time.Time) string {
	return t.In(time.UTC).Format(time.RFC3339)
}
