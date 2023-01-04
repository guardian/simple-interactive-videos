import boto3
import json


def lambda_handler(event, context):
    # Get the S3 bucket and key from the event
    bucket = event['Records'][0]['s3']['bucket']['name']
    key = event['Records'][0]['s3']['object']['key']
    
    s3 = boto3.client('s3')
    
    metadata = s3.head_object(Bucket=bucket, Key=key)
    
    print(metadata)
    
    # Set up SQS client
    sqs = boto3.client('sqs')

    queue_url = 'https://sqs.eu-west-1.amazonaws.com/855023211239/POC-interactives'

    # Send a message to the queue with the S3 bucket and key
    sqs.send_message(QueueUrl=queue_url, MessageBody=json.dumps({'bucket': bucket, 'key': key, 'metadata': {'email': metadata['Metadata'].get('email', 'default@email.com'), 'orientation': metadata['Metadata'].get('orientation', 'l') }}))