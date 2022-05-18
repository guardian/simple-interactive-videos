package main

import (
	"context"
	"errors"
	"flag"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/davecgh/go-spew/spew"
	"github.com/guardian/simple-interactive-deliverables/common"
	"log"
	"os"
	"time"
)

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
	params := dynamodb.QueryInput{
		TableName:                 tableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		IndexName:                 aws.String("filebase"),
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
GenerateNewRecord will create a new entry in the given id mapping table. The ContentId field is randomly generated, and this
function will automatically retry if the given ID already exists until one is found that does not.
*/
func GenerateNewRecord(ddbClient *dynamodb.Client, tableName *string, filebase *string, maybeProject *string, maybeOctId *int64) (*common.IdMappingRecord, error) {
	newRecord := &common.IdMappingRecord{
		ContentId:  common.GenerateNumericId(),
		Filebase:   *filebase,
		Project:    maybeProject,
		Lastupdate: time.Now(),
		Octopus_id: maybeOctId,
	}

	params := dynamodb.PutItemInput{
		Item:                newRecord.ToDynamoRecord(),
		TableName:           tableName,
		ConditionExpression: aws.String("attribute_not_exists(id)"), //see https://stackoverflow.com/questions/55106784/how-to-insert-to-dynamodb-just-if-the-key-does-not-exist
	}
	_, err := ddbClient.PutItem(context.Background(), &params)

	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() == "ConditionalCheckFailedException" {
				log.Printf("DEBUG Detected ID conflict on generated number %d, re-trying", newRecord.ContentId)
				return GenerateNewRecord(ddbClient, tableName, filebase, maybeProject, maybeOctId)
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

func main() {
	filebase := flag.String("filebase", "", "filebase for the title")
	octid := flag.Int64("octid", 0, "'octopus' (title) id (optional)")
	project := flag.String("project", "", "optional project ID")
	tableName := flag.String("table", "", "table name to put to")
	flag.Parse()

	if *filebase == "" {
		log.Fatal("You must specify a filebase to create a new title")
	}

	var maybeProject *string
	if *project != "" {
		maybeProject = project
	}

	var maybeOctId *int64
	if *octid != 0 {
		maybeOctId = octid
	}

	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("Could not connect to AWS: ", err)
	}
	ddbClient := dynamodb.NewFromConfig(awsConfig)

	existingRecordsCount, knownIdList, err := CheckForExistingName(ddbClient, tableName, filebase)
	if existingRecordsCount > 1 {
		log.Printf("WARNING: There are %d content IDs for the name '%s'. This will lead to problems and should be fixed.", existingRecordsCount, *filebase)
		for i, id := range knownIdList {
			log.Printf("ContentID %d: %s", i+1, id)
		}
		os.Exit(1)
	} else if existingRecordsCount > 0 {
		log.Printf("There is already a title for '%s' with this content ID: %s", *filebase, knownIdList[0])
		os.Exit(0)
	} else {
		newRecord, err := GenerateNewRecord(ddbClient, tableName, filebase, maybeProject, maybeOctId)
		if newRecord != nil {
			log.Print("Created new title record: ")
			spew.Dump(*newRecord)
		}

		if err == nil {
			log.Print("Title created")
		} else {
			log.Fatal("Could not create title: ", err)
		}
	}
}
