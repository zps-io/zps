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
	_, err := d.LockWithEtag()
	return err
}

// Unlock is used when we need to unlock repo but metadata wasn't changed
func (d *DynamoLocker) Unlock() error {
	_, err := d.client.UpdateItem(&dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(d.key)},
		},
		UpdateExpression: aws.String("SET Locked = :false"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":false": {BOOL: aws.Bool(false)},
		},
		TableName: &d.tableName,
	})
	return err

}

func (d *DynamoLocker) LockWithEtag() (string, error) {
	eTag := ""

	result, err := d.client.UpdateItem(&dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(d.key)},
		},
		UpdateExpression: aws.String("SET Locked = :true"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":true":  {BOOL: aws.Bool(true)},
			":false": {BOOL: aws.Bool(false)},
		},
		ConditionExpression: aws.String("attribute_not_exists(Locked) or Locked = :false"),
		ReturnValues:        aws.String("ALL_NEW"),
		TableName:           &d.tableName,
	})

	if err != nil {
		return eTag, err
	}

	if eTagResult, ok := result.Attributes["ETag"]; ok {
		eTag = *eTagResult.S
	}
	return eTag, nil
}

func (d *DynamoLocker) UnlockWithEtag(eTag *string) error {
	_, err := d.client.UpdateItem(&dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(d.key)},
		},
		UpdateExpression: aws.String("SET Locked = :false, ETag = :eTag"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":false": {BOOL: aws.Bool(false)},
			":eTag":  {S: eTag},
		},
		TableName: &d.tableName,
	})

	if err != nil {
		return err
	}

	return nil
}
