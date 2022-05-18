package main

import (
	"context"
	"flag"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/davecgh/go-spew/spew"
	"log"
	"os"
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
		newRecord, err := GenerateNewRecord(ddbClient, tableName, filebase, maybeProject, maybeOctId, 0)
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
