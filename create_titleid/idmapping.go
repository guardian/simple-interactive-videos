package main

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/guardian/simple-interactive-deliverables/common"
	"log"
	"time"
)

func IndexLookup(ddbClient *dynamodb.Client, tableName *string, index string, expr *expression.Expression) (int32, []string, error) {
	params := dynamodb.QueryInput{
		TableName:                 tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		IndexName:                 aws.String(index),
		KeyConditionExpression:    expr.KeyCondition(),
	}
	response, err := ddbClient.Query(context.Background(), &params)
	if err != nil {
		return -1, nil, err
	}

	knownIdList := make([]string, 0)
	for _, r := range response.Items {
		if contentId, haveContentId := r["contentid"]; haveContentId {
			if numericValue, isNumeric := contentId.(*types.AttributeValueMemberN); isNumeric {
				knownIdList = append(knownIdList, numericValue.Value)
			}
		}
	}
	return response.Count, knownIdList, nil
}

/*
CheckForExistingName returns the number of matching records for the given filebase.  Only if this
function returns zero should you continue to create a record
*/
func CheckForExistingName(ddbClient *dynamodb.Client, tableName *string, filebase *string) (int32, []string, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("filebase").Equal(expression.Value(*filebase))).
		Build()
	if err != nil {
		return -1, nil, err
	}

	return IndexLookup(ddbClient, tableName, "filebase", &expr)
}

//Note - needs a schema update

/*
CheckForExistingContentId returns the number of matching records for the given ContentId value.  Only if this
function returns zero should you continue to create a record
*/
func CheckForExistingContentId(ddbClient *dynamodb.Client, tableName *string, contentId int32) (int32, []string, error) {
	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("contentid").Equal(expression.Value(contentId))).
		Build()
	if err != nil {
		return -1, nil, err
	}

	return IndexLookup(ddbClient, tableName, "contentid", &expr)
}

/*
GenerateNewRecord will create a new entry in the given id mapping table. The ContentId field is randomly generated, and this
function will automatically retry if the given ID already exists until one is found that does not.
*/
func GenerateNewRecord(ddbClient *dynamodb.Client, tableName *string, filebase *string, maybeProject *string, maybeOctId *int64, attempt int, contentId int32) (*common.IdMappingRecord, error) {
	if attempt > 100 {
		return nil, errors.New("could not create an ID after 100 attempts, giving up")
	}

	newRecord := &common.IdMappingRecord{
		ContentId:  contentId,
		Filebase:   *filebase,
		Project:    maybeProject,
		Lastupdate: time.Now(),
		Octopus_id: maybeOctId,
	}

	newRecord.RegenerateUUID()

	//we will only create a record if there is not a pre-existing primary key
	condition, err := expression.NewBuilder().WithCondition(expression.AttributeNotExists(expression.Name("uuid"))).Build()
	if err != nil {
		return nil, err
	}

	params := dynamodb.PutItemInput{
		Item:                      newRecord.ToDynamoRecord(),
		TableName:                 tableName,
		ConditionExpression:       condition.Condition(),
		ExpressionAttributeValues: condition.Values(),
		ExpressionAttributeNames:  condition.Names(),
	}
	_, err = ddbClient.PutItem(context.Background(), &params)

	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "ConditionalCheckFailedException" {
				log.Printf("DEBUG Detected ID conflict on generated number %d, re-trying", newRecord.ContentId)
				return GenerateNewRecord(ddbClient, tableName, filebase, maybeProject, maybeOctId, attempt+1, contentId)
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return newRecord, nil
	}
}
