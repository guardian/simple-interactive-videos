package common

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/elastictranscoder/types"
	"github.com/aws/smithy-go/rand"
	"log"
	"os"
	"strconv"
	"time"
)

/*
Encoding represents a row from the `Encodings` table
*/
type Encoding struct {
	EncodingId  int32     `json:"encoding_id"` //NOT NULL
	ContentId   int32     `json:"content_id"`  //NOT NULL
	Url         string    `json:"url"`         //NOT NULL
	Format      string    `json:"format"`      //NOT NULL
	Mobile      bool      `json:"mobile"`      //NOT NULL
	Multirate   bool      `json:"multirate"`   //NOT NULL
	VCodec      string    `json:"vcodec"`
	ACodec      string    `json:"acodec"`
	VBitrate    int32     `json:"vbitrate"`
	ABitrate    int32     `json:"abitrate"`
	LastUpdate  time.Time `json:"last_update"`  //NOT NULL, defaults to current time
	FrameWidth  int32     `json:"frame_width"`  //NOT NULL
	FrameHeight int32     `json:"frame_height"` //NOT NULL
	Duration    float32   `json:"duration"`     //NOT NULL
	FileSize    int64     `json:"file_size"`    //NOT NULL
	FCSID       string    `json:"fcs_id"`       //NOT NULL
	OctopusId   int32     `json:"octopus_id"`   //NOT NULL aka 'title id'
	Aspect      string    `json:"aspect"`       //NOT NULL
}

/*
generateNumericId generates a (theoretically) unique numeric ID
*/
func GenerateNumericId() int32 {
	content, err := GenerateUuidBytes()
	if err != nil {
		log.Fatal("Could not get uuid bytes: ", err)
	}
	intval, _ := binary.Uvarint(content[0:8])
	return int32(intval) + 9900000
}

func GenerateStringId() string {
	content, err := GenerateUuidBytes()
	if err != nil {
		log.Fatal("Could not get uuid bytes: ", err)
	}
	return base64.StdEncoding.EncodeToString(content)
}

func GenerateUuidBytes() ([]byte, error) {
	entropy, err := os.Open("/dev/urandom")
	if err != nil {
		log.Fatal("Could not get entropy generator: ", err)
	}
	defer entropy.Close()
	u := rand.NewUUID(entropy)
	return u.GetBytes()
}

func GenerateUuidString() (string, error) {
	entropy, err := os.Open("/dev/urandom")
	if err != nil {
		log.Fatal("Could not get entropy generator: ", err)
	}
	defer entropy.Close()
	u := rand.NewUUID(entropy)
	return u.GetUUID()
}

func JobOutputToEncoding(o *types.JobOutput, presetInfo *types.Preset, contentId int32, titleId int32, urlBase string) *Encoding {
	encodingUrl := fmt.Sprintf("%s/%s", urlBase, *o.Key)

	//log.Printf("Output file %s has format %s, vcodec %s, acodec %s, vbitrate %s, abitrate %s",
	//	*o.Key,
	//	*presetInfo.Container,
	//	*presetInfo.Video.Codec,
	//	*presetInfo.Audio.Codec,
	//	*presetInfo.Video.BitRate,
	//	*presetInfo.Audio.BitRate,
	//	)

	formatMajor := "audio" //if the "video" part of the preset is configured we switch it to "video"
	var bitRateNum int32
	if videoPreset := presetInfo.Video; videoPreset != nil {
		temp, err := strconv.ParseInt(*videoPreset.BitRate, 10, 32)
		if err != nil {
			log.Fatalf("Could not convert bitrate string '%s' to number: %s", *videoPreset.BitRate, err)
		}
		formatMajor = "video"
		bitRateNum = int32(temp)
	}

	var audBitRateNum int32
	if audioPreset := presetInfo.Audio; audioPreset != nil {
		temp, err := strconv.ParseInt(*audioPreset.BitRate, 10, 32)
		if err != nil {
			log.Fatalf("Could not convert bitrate string '%s' to number: %s", *audioPreset.BitRate, err)
		}
		audBitRateNum = int32(temp)
	}

	var maybeAspectRatio string
	if aspect := presetInfo.Video.AspectRatio; aspect != nil {
		maybeAspectRatio = *aspect
	} else {
		maybeAspectRatio = ""
	}

	formatString := fmt.Sprintf("%s/%s", formatMajor, *presetInfo.Container)

	return &Encoding{
		EncodingId:  GenerateNumericId(),
		ContentId:   contentId,
		Url:         encodingUrl,
		Format:      formatString,
		Mobile:      false,
		Multirate:   false,
		VCodec:      *presetInfo.Video.Codec,
		ACodec:      *presetInfo.Audio.Codec,
		VBitrate:    bitRateNum,
		ABitrate:    audBitRateNum,
		LastUpdate:  time.Now(),
		FrameWidth:  *o.Width,
		FrameHeight: *o.Height,
		Duration:    float32(*o.Duration),
		FileSize:    *o.FileSize,
		FCSID:       GenerateStringId(),
		OctopusId:   titleId,
		Aspect:      maybeAspectRatio,
	}
}
