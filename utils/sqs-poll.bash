#!/bin/bash

export AWS_REGION="eu-west-1"

AWS_S3_BUCKET=""
MAPPING_TABLE=""
ENCODING_TABLE=""
CDN_BUCKET=""
PIPELINE_ID=""
URI_BASE=""
ENDPOINT=""

# Set the queue URL
queue_url=

while true; do
  # Retrieve a single message from the queue, waiting up to 1 minute for a message to become available
  message=$(aws sqs receive-message --region eu-west-1 --queue-url $queue_url --max-number-of-messages 10 --wait-time-seconds 20)

  # If there are no messages, sleep for 2 minutes
  if [ -z "$message" ]; then
    sleep 20
    continue
  fi

  # Print the metadata of the message
  body=$(echo $message | jq -r '.Messages[].Body')
  receipt_handle=$(echo $message | jq -r '.Messages[].ReceiptHandle')
  message_id=$(echo $message | jq -r '.Messages[].MessageId')
  echo "Message ID: $message_id"
  echo "Receipt Handle: $receipt_handle"
  echo "Body: $body"
  echo
  file=$(echo $body | jq .key | tr -d '"')
  file_id=$(echo $(echo $file | tr -d '"' | awk -F. '{$NF=""; print $0}' | tr ' ' '.' | rev | cut -c 2- | rev)$(echo $RANDOM | md5sum | head -c 5;))

  echo File: $file
  echo File ID: $file_id

  CONTENT_ID=$(./create_titleid.linux_x86 \
    -table ${MAPPING_TABLE} \
    -filebase $file_id \
    | grep ContentId \
    | awk '{print $NF}' \
    | tr "," " ")

  echo Content ID: $CONTENT_ID
  sleep 5

  ./transcodelauncher.linux_x86 \
    -input $file \
    -pipeline ${PIPELINE_ID} \
    -transcodeset horizonal_transcode_set.yaml \
    -contentId ${CONTENT_ID} \
    -table ${ENCODING_TABLE} \
    -cdnbucket ${CDN_BUCKET}:/$file_id \
    -uribase ${URI_BASE}$file_id

  echo Job URL: ${ENDPOINT}$file_id

  echo Deleting message: $receipt_handle
  aws sqs delete-message --region eu-west-1 --queue-url $queue_url --receipt-handle $receipt_handle
  # Sleep for 10 seconds before checking the queue again
  sleep 10
done