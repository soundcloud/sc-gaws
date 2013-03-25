// Types and functions to push metrics to AWS Cloudwatch
package main

import (
	"./credentials"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
    "stats"
	"sort"
	"strings"
	"time"
)

const (
	cloudwatchEndpoint   = "https://monitoring.amazonaws.com/doc/2010-08-01"
	cloudwatchApiVersion = "2010-08-01"
	maxMetricsPerRequest = 20
)

var b64 = base64.StdEncoding

// AwsStatsPusher implements the StatsPusher interface to allow
// pushing metrics to AWS Cloudwatch
type AwsStatsPusher struct {
	// AWS Credentials
	Credentials credentials.CredentialsProvider

	// Namespace for the metric e.g. bobone-cluster1, bobone-cluster2
	Namespace string
}

// Push a slice of metrics to CloudWatch
func (p AwsStatsPusher) Push(metrics []stats.Metric) {
	// make multiple requests to CloudWatch to send all the metrics
	n := len(metrics) / maxMetricsPerRequest
	if (len(metrics) % maxMetricsPerRequest) > 0 {
		n = n + 1
	}
	for i := 0; i < n; i++ {
		var m []stats.Metric
		if i == n-1 {
			m = metrics[(i * maxMetricsPerRequest):]
		} else {
			m = metrics[(i * maxMetricsPerRequest):((i + 1) * maxMetricsPerRequest)]
		}
		urlStr := fmt.Sprintf("%s/?Action=PutMetricData&Version=%s&Namespace=%s&Timestamp=%s&%s", cloudwatchEndpoint, cloudwatchApiVersion, p.Namespace, url.QueryEscape(timeInRfc3339(time.Now())), marshalMetrics(m))
		req, _ := http.NewRequest("GET", urlStr, nil)
		sign(req, p.Credentials.GetCredentials())
		client := http.Client{}
		log.Printf("Pushing metrics: %s", req.URL)
		res, err := client.Do(req)
		if err != nil {
			log.Printf("Pushing metrics failed with error: %s", err)
		} else {
			defer res.Body.Close()
			if res.StatusCode != 200 {
				body, _ := ioutil.ReadAll(res.Body)
				log.Printf("Pushing metrics failed with status code %d: %s", res.StatusCode, body)
			}
		}
	}
}

// Marshal metrics into a query string format for AWS Cloudwatch
func marshalMetrics(metrics []stats.Metric) string {
	metricsStr := make([]string, len(metrics))
	for i, m := range metrics {
		metricsStr[i] = fmt.Sprintf("MetricData.member.%d.MetricName=%s&MetricData.member.%d.Value=%f&MetricData.member.%d.Timestamp=%s", i+1, m.Name, i+1, m.Value, i+1, url.QueryEscape(timeInRfc3339(m.Timestamp)))
	}

	return strings.Join(metricsStr, "&")
}

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
