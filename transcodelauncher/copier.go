package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"log"
	"regexp"
)

type Copier struct {
	awsConfig aws.Config
	client    *s3.Client
}

func NewCopier(config aws.Config) *Copier {
	client := s3.NewFromConfig(config)
	return &Copier{awsConfig: config, client: client}
}

func (c *Copier) DoCopy(ctx context.Context, sourceBucket string, sourceKey string, destBucket string, destKey string, public bool) error {
	headReq := &s3.HeadObjectInput{
		Bucket: aws.String(sourceBucket),
		Key:    aws.String(sourceKey),
	}

	headResponse, err := c.client.HeadObject(ctx, headReq)
	if err != nil {
		return err
	}

	acl := types.ObjectCannedACLPrivate
	if public {
		acl = types.ObjectCannedACLPublicRead
	}

	log.Printf("Copying from s3://%s/%s to s3://%s/%s with %s ACL", sourceBucket, sourceKey, destBucket, destKey, acl)

	copyReq := &s3.CopyObjectInput{
		Bucket:      aws.String(destBucket),
		CopySource:  aws.String(fmt.Sprintf("%s/%s", sourceBucket, sourceKey)),
		Key:         aws.String(destKey),
		ACL:         acl,
		ContentType: headResponse.ContentType,
	}
	_, err = c.client.CopyObject(ctx, copyReq)
	if err != nil {
		return err
	}
	return nil
}

func (c *Copier) DoCopyDestspec(ctx context.Context, sourceBucket string, sourceKey string, destSpec string, public bool) error {
	splitter := regexp.MustCompile("([^:]+):/*(.*)")
	parts := splitter.FindAllStringSubmatch(destSpec, -1)
	if parts == nil {
		return errors.New(fmt.Sprintf("destSpec %s is malformed", destSpec))
	}

	return c.DoCopy(ctx, sourceBucket, sourceKey, parts[0][1], parts[0][2], public)
}

func (c *Copier) DoesFileExist(ctx context.Context, bucket string, key string) (bool, error) {
	req := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	log.Printf("Checking for existence of s3://%s/%s", bucket, key)
	_, err := c.client.HeadObject(ctx, req)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "NotFound" {
				return false, nil
			} else {
				log.Printf("ERROR Could not check file s3://%s/%s %s %s", bucket, key, ae.ErrorCode(), ae.ErrorMessage())
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}
