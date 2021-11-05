package zpm

import (
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type DynamoLocker struct {
	client    *dynamodb.DynamoDB
	tableName string
	key       string
}

func NewDynamoLocker(uri *url.URL) *DynamoLocker {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	tableName := uri.Host
	key := uri.Path

	return &DynamoLocker{
		client:    svc,
		tableName: tableName,
		key:       key,
	}
}

func (d *DynamoLocker) Lock() error {
	_, err := d.client.PutItem(&dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(d.key)},
		},
		TableName:           &d.tableName,
		ConditionExpression: aws.String("attribute_not_exists(LockKey)"),
	})

	if err != nil {
		return err
	}
	return nil
}

func (d *DynamoLocker) Unlock() error {
	_, err := d.client.DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(d.key)},
		},
		TableName: &d.tableName,
	})

	if err != nil {
		return err
	}

	return nil
}
