package main

import (
	"context"
	"flag"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/davecgh/go-spew/spew"
	"github.com/guardian/simple-interactive-deliverables/common"
	"log"
	"time"
)

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

	newRecord := &common.IdMappingRecord{
		ContentId:  common.GenerateNumericId(),
		Filebase:   *filebase,
		Project:    maybeProject,
		Lastupdate: time.Now(),
		Octopus_id: maybeOctId,
	}

	log.Print("Creating new title record: ")
	spew.Dump(*newRecord)

	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("Could not connect to AWS: ", err)
	}
	ddbClient := dynamodb.NewFromConfig(awsConfig)

	params := dynamodb.PutItemInput{
		Item:      newRecord.ToDynamoRecord(),
		TableName: tableName,
	}
	_, err = ddbClient.PutItem(context.Background(), &params)
	if err == nil {
		log.Print("Title created")
	} else {
		log.Fatal("Could not create title: ", err)
	}
}
