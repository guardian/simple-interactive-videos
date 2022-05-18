package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"time"
)

type IdMappingRecord struct {
	UUID       string
	ContentId  int32
	Filebase   string //base index
	Project    *string
	Lastupdate time.Time //range key for all indices
	Octopus_id *int64    //indexed
}

func NewIdMappingRecord() *IdMappingRecord {
	uidStr, err := GenerateUuidString()
	if err != nil {
		log.Fatal("Could not generate uuid: ", err)
	}

	return &IdMappingRecord{UUID: uidStr}
}

func (r *IdMappingRecord) RegenerateUUID() {
	uidStr, err := GenerateUuidString()
	if err != nil {
		log.Fatal("Could not generate uuid: ", err)
	}
	r.UUID = uidStr
}

func (r *IdMappingRecord) ToDynamoRecord() map[string]types.AttributeValue {
	var maybeProjectAttr types.AttributeValue
	if p := r.Project; p != nil {
		maybeProjectAttr = &types.AttributeValueMemberS{Value: *r.Project}
	} else {
		maybeProjectAttr = &types.AttributeValueMemberNULL{Value: true}
	}
	var maybeOctId types.AttributeValue
	if o := r.Octopus_id; o != nil {
		maybeOctId = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", o)}
	} else {
		maybeOctId = &types.AttributeValueMemberNULL{Value: true}
	}

	return map[string]types.AttributeValue{
		"uuid":       &types.AttributeValueMemberS{Value: r.UUID},
		"contentid":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", r.ContentId)},
		"filebase":   &types.AttributeValueMemberS{Value: r.Filebase},
		"project":    maybeProjectAttr,
		"lastupdate": &types.AttributeValueMemberS{Value: r.Lastupdate.Format(time.RFC3339)},
		"octopusId":  maybeOctId,
	}
}
