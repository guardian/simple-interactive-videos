package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/guardian/simple-interactive-deliverables/common"
	"log"
	"os"
	"path"
	"regexp"
	"time"
)

/*
isCompleted returns true if the given string pointer refers to a completed status
(i.e. success/failure/cancelled)
*/
func isCompleted(jobStatus *string) bool {
	return *jobStatus == "Complete" || *jobStatus == "Canceled" || *jobStatus == "Error"
}

func getPresetInfo(cli *elastictranscoder.Client, presetId string) *elastictranscoder.ReadPresetOutput {
	out, err := cli.ReadPreset(context.Background(), &elastictranscoder.ReadPresetInput{Id: aws.String(presetId)})
	if err != nil {
		log.Fatalf("Could not read preset %s: %s", presetId, err)
	}
	return out
}

func appendToBasefile(originalFilename *string, stuff string) string {
	xtractor := regexp.MustCompile("(.*)\\.([^.]*)")

	fileBase := path.Base(*originalFilename)
	var fileStem string
	parts := xtractor.FindAllStringSubmatch(fileBase, -1)
	if parts != nil {
		fileStem = parts[0][1]
	} else {
		fileStem = fileBase
	}
	return fileStem + stuff
}

func makeOutputFilename(originalFilename *string, transcode *TranscodeSet) string {
	return appendToBasefile(originalFilename, transcode.Suffix+transcode.Extension)
}

func makeOutputThumbsPattern(originalFilename *string, transcode *TranscodeSet) string {
	return appendToBasefile(originalFilename, fmt.Sprintf("_{count}%s", transcode.Suffix))
}

func WriteOutputs(awsConfig aws.Config, tableName string, content []*common.Encoding) {
	client := dynamodb.NewFromConfig(awsConfig)

	for _, entry := range content {
		toWrite := entry.ToDynamoDB()
		_, err := client.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:      toWrite,
			TableName: &tableName,
		})
		if err != nil {
			log.Fatal("Could not write item: ", err)
		}
	}
}

func main() {
	inputFile := flag.String("input", "", "Bucket path to input file")
	overridePrefix := flag.String("outputprefix", "", "Prefix for intermediate output files. If empty, a randomly generated prefix is used")
	pipelineId := flag.String("pipeline", "", "Pipeline ID to run on")
	transcodeSet := flag.String("transcodeset", "horizontal_transcode_set", "transcode set yaml to use")
	contentId := flag.Int64("contentId", 0, "content ID value to use")
	titleId := flag.Int64("titleId", 0, "title ID value to use")
	uriBase := flag.String("uribase", "", "base URL on the CDN where content is accessible")
	tableName := flag.String("table", "", "name of the table that contains encodings")
	noDbOut := flag.Bool("nodb", false, "set this to only run the transcode and not push to encodings endpoint")
	cdnPath := flag.String("cdnbucket", "", "If set, copy the files to this location in the form bucket:/path")
	flag.Parse()

	transcodes, err := LoadTranscodeSet(transcodeSet)
	if err != nil {
		log.Fatalf("Could not load transcodes from '%s': '%s'", *transcodeSet, err)
	}

	if *tableName == "" {
		log.Fatal("You must specify a table to output encodings to")
	}

	var outputPrefix string
	if overridePrefix == nil {
		outputPrefix = *overridePrefix
	} else {
		outputPrefix = common.GenerateStringIdPathSafe() + "/"
	}

	awscfg, awsErr := awsconfig.LoadDefaultConfig(context.Background())
	if awsErr != nil {
		log.Fatalf("Could not connect to AWS: %s", awsErr)
	}

	cli := elastictranscoder.NewFromConfig(awscfg)

	query := &elastictranscoder.ReadPipelineInput{Id: pipelineId}
	pipelineInfo, err := cli.ReadPipeline(context.Background(), query)
	if err != nil {
		log.Fatalf("Could not query pipeline %s: %s", *pipelineId, err)
	}
	log.Printf("INFO Pipeline %s runs from %s into %s; status %s", *pipelineId, *pipelineInfo.Pipeline.InputBucket, *pipelineInfo.Pipeline.OutputBucket, *pipelineInfo.Pipeline.Status)

	outputList := make([]types.CreateJobOutput, len(*transcodes))
	for i, transcode := range *transcodes {
		outputList[i] = types.CreateJobOutput{
			Key:              aws.String(makeOutputFilename(inputFile, &transcode)),
			PresetId:         aws.String(transcode.PresetId),
			ThumbnailPattern: aws.String(makeOutputThumbsPattern(inputFile, &transcode)),
		}
	}

	params := &elastictranscoder.CreateJobInput{
		PipelineId: aws.String(*pipelineId),
		Input: &types.JobInput{
			Key: aws.String(*inputFile),
		},
		OutputKeyPrefix: aws.String(outputPrefix),
		Outputs:         outputList,
		Playlists:       nil,
		UserMetadata:    nil,
	}

	response, err := cli.CreateJob(context.Background(), params)
	if err != nil {
		log.Fatalf("Could not start ETS job for '%s': %s", *inputFile, err)
	}

	log.Printf("Created job with ID %s", *response.Job.Id)

	var jobStatus *elastictranscoder.ReadJobOutput
	for {
		time.Sleep(5 * time.Second)
		jobStatus, err = cli.ReadJob(context.Background(), &elastictranscoder.ReadJobInput{Id: response.Job.Id})
		if err != nil {
			log.Fatalf("Could not check job status: %s", err)
		}

		log.Printf("Job status is %s", *jobStatus.Job.Status)
		if isCompleted(jobStatus.Job.Status) {
			break
		}
	}

	fcsId := common.GenerateStringId()

	copier := NewCopier(awscfg)

	switch *jobStatus.Job.Status {
	case "Error":
		log.Printf("Job failed, please see the ETS console for job ID '%s' to get the reason", *jobStatus.Job.Id)
		os.Exit(1)
	case "Canceled":
		log.Printf("Job was cancelled in the ETS console")
		os.Exit(2)
	case "Complete":
		log.Printf("Job completed!")
		enc := make([]*common.Encoding, len(jobStatus.Job.Outputs))
		for i, out := range jobStatus.Job.Outputs {
			presetInfo := getPresetInfo(cli, *out.PresetId)
			enc[i] = common.JobOutputToEncoding(&out, presetInfo.Preset, int32(*contentId), int32(*titleId), fcsId, *uriBase)
			spew.Dump(enc[i])
			outputFilepath := *out.Key

			if outputPrefix != "" {
				outputFilepath = outputPrefix + *out.Key
			}
			if *cdnPath != "" {
				currentPosterName, destPosterName, err := PosterFrameNamesForEncoding(outputFilepath, ".png")

				log.Printf("CurrentPosterName is %s from %s", currentPosterName, outputFilepath)
				havePoster, err := copier.DoesFileExist(context.Background(), *pipelineInfo.Pipeline.OutputBucket, currentPosterName)
				if err != nil {
					log.Fatal("Could not locate proxy: ", err)
				}
				if havePoster {
					err = copier.DoCopyDestspec(context.Background(),
						*pipelineInfo.Pipeline.OutputBucket,
						currentPosterName,
						*cdnPath+"/"+destPosterName,
						true)

					if err != nil {
						log.Printf("ERROR Could not copy poster s3://%s/%s: %s", *pipelineInfo.Pipeline.OutputBucket, currentPosterName, err)
					}
				} else {
					log.Printf("WARNING Could not find a poster for %s", *out.Key)
				}
				err = copier.DoCopyDestspec(context.Background(),
					*pipelineInfo.Pipeline.OutputBucket,
					outputFilepath,
					*cdnPath+"/"+*out.Key,
					true)

				if err != nil {
					log.Printf("ERROR Could not copy s3://%s/%s: %s", *pipelineInfo.Pipeline.OutputBucket, outputFilepath, err)
				}
			}
		}
		if !*noDbOut {
			WriteOutputs(awscfg, *tableName, enc)
		}
	default:
		log.Fatalf("Got an unexpected job status: '%s'", *jobStatus.Job.Status)
	}
}
