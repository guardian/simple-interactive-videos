# Simple interactive videos POC

This POC triggers an automated process when a video is upload to the interavtives s3 input bucket.

The `s3-trigger-lambda.py` lambda function creates an SQS message with the file name and metadata for the video orientation and users email.

This message is picked up by an ec2 instance running the `sqs-poll.bash` script which runs the `create_titleid` and `transcodelauncher` binaries.

## To use:
Edit these lines so that the script points to the correct resources.
```AWS_S3_BUCKET=""
MAPPING_TABLE=""
ENCODING_TABLE=""
CDN_BUCKET=""
PIPELINE_ID=""
URI_BASE=""
```

The video file can be uploaded with optional metadata from the command line like this:
```
$ aws s3 cp my-test-file.mp4 s3://my-input-bucket --metadata '{"email":"test@example.com", "orientation":"p"}'
```


## TODO:

- Send the user an email containing urls to the transcoded files.



