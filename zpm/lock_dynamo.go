package zpm

import (
	"encoding/hex"
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

func (d *DynamoLocker) Unlock() error {
	eTag := [16]byte{}
	return d.UnlockWithEtag(eTag)
}

func (d *DynamoLocker) LockWithEtag() ([16]byte, error) {
	var eTag [16]byte

	result, err := d.client.UpdateItem(&dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(d.key)},
		},
		UpdateExpression: aws.String("SET Locked = :true"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":true":  {BOOL: aws.Bool(true)},
			":false": {BOOL: aws.Bool(false)},
		},
		ConditionExpression: aws.String("Locked = :false"),
		ReturnValues:        aws.String("ETag"),
		TableName:           &d.tableName,
	})
	if err != nil {
		return eTag, err
	}

	if len(result.String()) > 0 {
		eTagSlice, err := hex.DecodeString(result.String())
		if err != nil {
			return eTag, err
		}

		copy(eTag[:], eTagSlice)

	}
	return eTag, nil
}

func (d *DynamoLocker) UnlockWithEtag(eTag [16]byte) error {
	eTagString := hex.EncodeToString(eTag[:])
	_, err := d.client.UpdateItem(&dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockKey": {S: aws.String(d.key)},
		},
		UpdateExpression: aws.String("SET Locked = :false, ETag = :eTag"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":false": {BOOL: aws.Bool(false)},
			":eTag":  {S: aws.String(eTagString)},
		},
		TableName: &d.tableName,
	})

	if err != nil {
		return err
	}

	return nil
}
