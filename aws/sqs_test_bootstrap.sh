#!/bin/bash

echo 'Usage: "source sqs_test_bootstrap.sh" to set up environment vars for SQS test'

QUEUE_NAME_PREFIX="HlsdStaging"
export SQS_TEST_ENDPOINT=$(aws sqs list-queues --queue-name-prefix $QUEUE_NAME_PREFIX --query QueueUrls[0] | tr -d \")
