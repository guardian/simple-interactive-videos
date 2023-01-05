# Simple interactive videos POC

This POC triggers an automated process when a video is upload to the interavtives s3 input bucket.

The `s3-trigger-lambda.py` lambda function creates an SQS message with the file name and metadata for the video orientation and users email.

This message is picked up by an ec2 instance running the `sqs-poll.bash` script which runs the `create_titleid` and `transcodelauncher` binaries.

## TODO:

- Send the user an email containing urls to the transcoded files.



