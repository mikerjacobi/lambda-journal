package journal

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Entry struct {
	EntryID string `json:"entry_id"`
	Entry   string `json:"entry"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

func (e Entry) GetItem() *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: aws.String("entries"),
		Key: map[string]*dynamodb.AttributeValue{
			"entry_id": {S: aws.String(e.EntryID)},
		},
	}
}

func (e Entry) GotItem(item *dynamodb.GetItemOutput) Entry {
	if itemEntry, ok := item.Item["entry"]; ok {
		e.Entry = *itemEntry.S
	}
	if itemEntry, ok := item.Item["created"]; ok {
		e.Created = *itemEntry.S
	}
	if itemEntry, ok := item.Item["updated"]; ok {
		e.Updated = *itemEntry.S
	}
	return e
}

func (e Entry) PutItem() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: aws.String("entries"),
		Item: map[string]*dynamodb.AttributeValue{
			"entry_id": {S: aws.String(e.EntryID)},
			"entry":    {S: aws.String(e.Entry)},
			"created":  {S: aws.String(e.Created)},
			"updated":  {S: aws.String(e.Updated)},
		},
	}
}
