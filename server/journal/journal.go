package journal

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Journal struct {
	JournalID string `json:"journal_id"`
	Entry     string `json:"entry"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
}

func (j Journal) GetItem() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: aws.String("journal"),
		Key: map[string]*dynamodb.AttributeValue{
			"journal_id": {S: aws.String(j.JournalID)},
		},
	}
}

func (j Journal) GotItem(item *dynamodb.GetItemOutput) Journal {
	if attr, ok := item.Item["entry"]; ok {
		j.Entry = *attr.S
	}
	if attr, ok := item.Item["created"]; ok {
		j.Created = *attr.S
	}
	if attr, ok := item.Item["updated"]; ok {
		j.Updated = *attr.S
	}
	return j
}

func (j Journal) PutItem() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: aws.String("journal"),
		Item: map[string]*dynamodb.AttributeValue{
			"journal_id": {S: aws.String(j.JournalID)},
			"entry":      {S: aws.String(j.Entry)},
			"created":    {S: aws.String(j.Created)},
			"updated":    {S: aws.String(j.Updated)},
		},
	}
}
