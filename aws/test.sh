#!/bin/bash

# Test suite for the AWS package

QUEUE_NAME_PREFIX="HlsdStaging"
export SQS_TEST_ENDPOINT=$(aws sqs list-queues --queue-name-prefix $QUEUE_NAME_PREFIX --query QueueUrls[0] | tr -d \")

go test