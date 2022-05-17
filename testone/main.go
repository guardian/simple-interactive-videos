package main

import (
	"context"
	"flag"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/guardian/simple-interactive-deliverables/common"
	"log"
	"os"
	"strings"
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

func main() {
	inputFile := flag.String("input", "", "Bucket path to input file")
	outputPrefix := flag.String("outputprefix", "", "Prefix for output files")
	pipelineId := flag.String("pipeline", "", "Pipeline ID to run on")
	presetIds := flag.String("presets", "", "Comma-separated list of presets to use")
	contentId := flag.Int64("contentId", 0, "content ID value to use")
	titleId := flag.Int64("titleId", 0, "title ID value to use")
	uriBase := flag.String("uribase", "", "base URL on the CDN where content is accessible")
	flag.Parse()

	presetIdList := strings.Split(*presetIds, ",")

	outputList := make([]types.CreateJobOutput, len(presetIdList))
	for i, presetId := range presetIdList {
		outputList[i] = types.CreateJobOutput{
			Key:      aws.String(*inputFile),
			PresetId: aws.String(presetId),
		}
	}
	params := &elastictranscoder.CreateJobInput{
		PipelineId: aws.String(*pipelineId),
		Input: &types.JobInput{
			Key: aws.String(*inputFile),
		},
		OutputKeyPrefix: aws.String(*outputPrefix),
		Outputs:         outputList,
		Playlists:       nil,
		UserMetadata:    nil,
	}

	awscfg, awsErr := awsconfig.LoadDefaultConfig(context.Background())
	if awsErr != nil {
		log.Fatalf("Could not connect to AWS: %s", awsErr)
	}

	cli := elastictranscoder.NewFromConfig(awscfg)

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
			enc[i] = common.JobOutputToEncoding(&out, presetInfo.Preset, int32(*contentId), int32(*titleId), *uriBase)
			spew.Dump(enc[i])
		}
	default:
		log.Fatalf("Got an unexpected job status: '%s'", *jobStatus.Job.Status)
	}
}
