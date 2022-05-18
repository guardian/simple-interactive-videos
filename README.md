# simple-interactive-videos

Or, the answer to the question "How do I put new content into the interactives endpoint?"

See https://github.com/guardian/new-mm-encodings-endpoints/ for information about the endpoints themselves

## Initial set up

### Get the binaries

On the releases page, you should find compiled binaries of all the tools.

There are versions for Mac, Windows and Linux; Mac and Linux are both compiled for ARM and x86_64.

Current binaries are:
 
- create_titleid - creates new titles
- transcodelauncher - runs transcodes, deploys media to CDN and writes to the database

If you want anything else, you'll have to download or clone the source code and run `make prod` to compile them yourself.

### Infrastructure

- Grab a copy of the file  `cloudformation/delivery-system.yaml`
- Deploy it. This will create a couple of buckets in your AWS account and some SNS topics for notifications
- Go to the Elastic Transcoder in the AWS Console in your browser
- Create a new pipeline. Select the "input" and "output" buckets that were just created by cloudformation, and (if you want)
  select the SNS topics to receive completion and failure messages

### Configure the encodings set

Before we can transcode anything, we need to configure the settings for transcoding into.  Review the available
presets in Elastic Transcoder and choose some to make up your "encodings set".

Go ahead and create more if you want.

If you want to have automatic poster frames, then make sure you set "generate thumbnails" as ON in your transcode preset
and set the thumbnail width/height to the same as the video width/height.

The default scaling, "shrink to fit", takes this width/height as maximum and scales the content appropriately. So
event 1920x1080 will work for a vertical video without stretching it.  The thumbnail should come out the same as the video.

Create a YAML document to hold it, in this format:

```yaml
- presetId: {copy the random number/letter ID into here}
  suffix: {this is something to append to the filename to uniquely identify the file. We expect it to start with a _}
  extension: {file extension that the file will get}
- presetId: {copy the random number/letter ID into here}
  suffix: {this is something to append to the filename to uniquely identify the file. We expect it to start with a _}
  extension: {file extension that the file will get}
- presetId: {copy the random number/letter ID into here}
  suffix: {this is something to append to the filename to uniquely identify the file. We expect it to start with a _}
  extension: {file extension that the file will get}
```

You can have as many blocks as you like in here.

There's an example one, `horizontal_transcode_set`, in the `transcodelauncher` directory already

## Step zero - make sure you have AWS credentials

These tools use commandline credentials in exactly the same way as the `aws` command.  So make sure that you can use that.

You specifically need read/write permissions on the following services:
- S3
- ElasticTranscoder
- DyanamoDB

## Step one - create a title ID

The first step is to create a "title", which all your transcodes will be linked to.  This is done by `create_titleid` as follows

```bash 
create_titleid -tableName {idmapping-table} -filebase {your-titlename} 
```

You need to get the correct value for `{idmapping-table}` from the actual endpoints deployment.
Consult the "resources" page of the endpoints' deployed Cloudformation and look for a DynamoDB table that has "idmapping"
in the name.

This will output a record, including a randomly generated "contentID". Save this, you'll need it for the next steps.

## Step two - upload your media

Assuming that you have the source media that you're going to transcode from locally, you need to put it into the "input" bucket
that you created via Cloudformation at the start:

```
aws s3 cp mymedia.mp4 s3://input-bucket-name
```

## Step three - transcode and output

These steps are automated by `transcodelauncher`. Run it as follows:

```bash
transcodelauncher -inputFile mymedia.mp4 -pipelineId {pipeline-id} -transcodeset transcodes.yaml -contentId {content-id} \
  -tableName {encodings-table} -cdnPath {cdn-bucket}:path/on/cdn -uriBase https://your-cdn-host/path/on/cdn [-noDbOut]
```

You need to specify quite a few parameters here:
- inputFile - this is the path to the _source media_ that you want to use, _on the "input" bucket_.
- pipelineId - this is the randomised number/letter ID of the Elastic Transcoder pipeline you set up at the start
- transcodeset - this is the yaml file that you created at the start, with your preset IDs in it
- contentId - this is the numeric value you got from `create_titleid`
- tableName - this is the name of the "Encodings" table from the actual endpoints deployment. Consult the "resources"
page of the endpoints' deployed Cloudformation and look for a DynamoDB table that has "encodings" in the name.
- cdnPath - optional. If specifed, the transcoded media files will all be copied from the output bucket to this S3 location.
Specify it in the form `bucketname:path/for/encodings`, so e.g. mymedia_1m.mp4 would go to the path `path/for/encodings/mymedia_1m.mp4` on
the bucket `bucketname`.
- uriBase - This is the location on the CDN where the content will be accessed. It's used to generate the https urls for
the endpoint.  You should set this up to be the correct https public-facing path to `cdnPath`.
- noDbOut - optional, for testing. If set, then the program won't attempt to output any encodings to the database.

## Step four - test

Having done this, you should be in a position to run:

```bash
curl -D- 'https://multimedia.guardianapis.com/interactivevideos/reference.php?file=your_title_name&format=video%2Fmp4'
```

And you should get back something looking like this:

```
HTTP/2 200 
content-type: text/plain;charset=UTF-8
content-length: 65
date: Wed, 18 May 2022 14:13:38 GMT
x-amzn-requestid: a24d2878-f5fc-4622-b75b-3ae5d49208b9
access-control-allow-origin: *
access-control-allow-headers: *
x-amzn-remapped-content-length: 65
x-amz-apigw-id: SUzK0HLJDoEF3kQ=
access-control-allow-methods: GET, OPTIONS
x-amzn-trace-id: Root=1-6284ff11-51f0e6cf5cfcb964370c8f8e;Sampled=0
access-control-max-age: 3600
access-control-allow-credentials: false
x-cache: Miss from cloudfront
via: 1.1 f735f4a6973fb5ea131811587853dcf6.cloudfront.net (CloudFront)
x-amz-cf-pop: LHR61-C2
x-amz-cf-id: xP4wlSY4pTmU_EdFCQJ6r8eF8eAcxIOwesc9bmXybQkdT1TZeVxCJA==

https://cdn.theguardian.tv/interactivevids/your_title_name_4m.mp4
```

Or:

```bash
curl -D- 'https://multimedia.guardianapis.com/interactivevideos/reference.php?file=your_title_name&format=video%2Fmp4&poster&png'
```

```
HTTP/2 200 
content-type: text/plain;charset=UTF-8
content-length: 72
date: Wed, 18 May 2022 14:14:37 GMT
x-amzn-requestid: aac9489d-ed4a-48cb-99dc-4573e9ad0257
access-control-allow-origin: *
access-control-allow-headers: *
x-amzn-remapped-content-length: 72
x-amz-apigw-id: SUzUJEYKjoEFTow=
access-control-allow-methods: GET, OPTIONS
x-amzn-trace-id: Root=1-6284ff4d-5e20c0ec63928cd85289f0b4;Sampled=0
access-control-max-age: 3600
access-control-allow-credentials: false
x-cache: Miss from cloudfront
via: 1.1 7cbc7be2814e4b470b205933b90a9fb0.cloudfront.net (CloudFront)
x-amz-cf-pop: LHR61-C2
x-amz-cf-id: kQOefH2rzrDXd_0ANyjceBtO3gp7t52-VLzZ1DRAIgXJ2hgAtKlrQg==

https://cdn.theguardian.tv/interactivevids/your_title_name_4m_poster.png
```