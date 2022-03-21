# ⚠️ Deprecation notice: Please note that this project has been deprecated and is no longer actively maintained. If you are interested in working on this and would like to take over ownership, please [get in touch with us here](mainto:developers@soundcloud.com)

---

# sc-gaws

Glue code to wrap around AWS and do useful things in Go.

## Usage

### aws/credentials

```
import (
    "github.com/kr/s3"
    "github.com/soundcloud/sc-gaws/aws/credentials"
)

func myFunc() {
    var credentialsProvider credentials.CredentialsProvider

    // With access key and secret key
    credentialsProvider = credentials.NewIamUserCredentials(accessKey, secretKey)

    // With role name
    var err error
    credentialsProvider, err = credentials.NewIamRoleCredentials(role)
    if err != nil {
        log.Fatalf("Error fetching credentials for role %s: %s", role, err)
    }

    // Create an HTTP client request
    // req, err := http.NewRequest(...)

    creds := credentialsProvider.GetCredentials()
    keys := s3.Keys{AccessKey: creds.AccessKeyId, SecretKey: creds.SecretAccessKey, SecurityToken: creds.Token}
    s3.Sign(req, keys)
}
```

### aws/elasticache

This package provides a mechanism for auto-discovery of ElastiCache servers.
It is recommended you get the configuration data on a 60 second timer and
keep track of configuration version numbers, so that you can set your
client to use the new set of servers only when they change.

```
import (
    "elasticache"
    "log"
)

var configVersion = 0

func updaterGoRoutine(autoDiscoveryConfigHost string) {
    autoDiscoverer, err := elasticache.NewAutoDiscoverer(autoDiscoveryConfigHost)
    if err != nil {
        log.Fatal(err.Error())
    }

    for {
        version, elastiCacheServers, err := autoDiscoverer.GetClusterConfig()

        if version > configVersion {
            configVersion = version

            // call SetServers on the Memcache client ServerList
        }

        time.Sleep(60 * time.Second)
    }
}
```

### aws and stats

```
import (
    "github.com/soundcloud/sc-gaws/aws"
    "github.com/soundcloud/sc-gaws/stats"
)

func myFunc() {
    metricsChan := make(chan stats.Metric)

    // You'll need to set up the credentials as per the above section.
    s := stats.NewStats(
        aws.AwsStatsPusher{
            credentialsProvider,
            "MyMetricNameSpace",
        },
        10, // number of samples to accumulate as one data point
    )

    go s.AccumulateAndPush(60*time.Second, metricsChan) // Push stats once every minute

    metricsChan <- stats.Metric{"NumberOfWidgetsCount", float32(1.0), "Count", time.Now()}

    begin := time.Now()
    // Some expensive operation
    duration := float32(time.Since(begin) / time.Millisecond)
    metricsChan <- stats.Metric{"WidgetResponseTimeMs", duration, "Average", time.Now()}
}
```

### aws/cloudfront

When using CloudFront with Restrict Viewer Access option, every URL needs to be signed.

This package implements:

- [Canned Policy signing](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-creating-signed-url-canned-policy.html).
- [Custom Policy signing](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-creating-signed-url-custom-policy.html).

_Please note:_ Due to performance reason we are using openssl c bindings.

It requires a valid [CloudFront private key](http://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-trusted-signers.html#private-content-creating-cloudfront-key-pairs).

```go
package main

import (
    "fmt"
    "time"

    "github.com/soundcloud/sc-gaws/aws/cloudfront"
)

func main() {
    privateKey, _ := cloudfront.NewRSAPrivateKeyFromFile("pk-KeyPairId.pem")

    baseURL := "http://cloudfront-foobar.com/my-sweet-file.mp3"
    wildcardURL := "*/my-sweet-file.mp3"
    expiresAt := time.Now().Add(time.Duration(2 * time.Minute))

    // Provide your URL and desired expiry time for a canned policy.
    cannedPolicy := cloudfront.NewCannedPolicy(baseURL, expiresAt)

    // Or provide your URL, an expiry time and an optional start time and
    // source IP address restriction for a custom policy.
    startsAt := time.Now()
    sourceIP := "192.0.2.1"
    customPolicy := cloudfront.NewCustomPolicy(wildcardURL, expiresAt).AddStartTime(startsAt).AddSourceIP(sourceIP)

    cannedSignature, _ := cloudfront.SignPolicy(privateKey, cannedPolicy)
    customSignature, _ := cloudfront.SignPolicy(privateKey, customPolicy)

    queryParams := "myparam=1"

    fmt.Printf("%s?%s&%s", baseURL, queryParams, cannedSignature)
    fmt.Printf("%s?%s&%s", baseURL, queryParams, customSignature)
}
```

Once you have generated a signature, you can append it to your base URL to
generate a working CloudFront URL based on the supplied parameters.
