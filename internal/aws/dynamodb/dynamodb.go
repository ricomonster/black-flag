package dynamodb

import (
	"context"
	"errors"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDB[T any] struct {
	client *dynamodb.Client
	table  string
}

// Instantiate and setups connectivity to DynamoDB and to the target table
func NewDynamoDB[T any](table string) *DynamoDB[T] {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	service := dynamodb.NewFromConfig(cfg)

	return &DynamoDB[T]{client: service, table: table}
}

// Returns a list of acceessible tables (this will vary per account iam role/permission)
func (ddb *DynamoDB[_]) ListTables() ([]string, error) {
	var tables []string
	var lastEvaluatedTableName *string
	for {
		resp, err := ddb.client.ListTables(context.TODO(), &dynamodb.ListTablesInput{
			ExclusiveStartTableName: lastEvaluatedTableName,
		})
		if err != nil {
			return []string{}, err
		}

		// Merge
		tables = append(tables, resp.TableNames...)

		// Check if there are more results to fetch.
		if resp.LastEvaluatedTableName == nil {
			// No more results, exit the loop.
			break
		}

		// Update the lastEvaluatedTableName to fetch the next page.
		lastEvaluatedTableName = resp.LastEvaluatedTableName
	}

	return tables, nil
}

// Find a record in the DynamoDB table using the primary key attribute value
func (ddb *DynamoDB[T]) FindById(key string, value string) (T, error) {
	var result T

	if key == "" {
		return result, errors.New("Key name is required.")
	}

	if value == "" {
		return result, errors.New("ID is required.")
	}

	findValue, err := attributevalue.Marshal(value)
	if err != nil {
		panic(err)
	}

	command := &dynamodb.GetItemInput{
		TableName: aws.String(ddb.table),
		Key:       map[string]types.AttributeValue{key: findValue},
	}

	response, err := ddb.client.GetItem(context.TODO(), command)
	if err != nil {
		log.Printf("Couldn't get info about %v. Here's why: %v\n", value, err)
		return result, err
	}

	err = attributevalue.UnmarshalMap(response.Item, &result)
	if err != nil {
		log.Printf("Couldn't unmarshal response. Here's why: %v\n", err)
		return result, err
	}

	return result, err
}

// Inserts an item to a DynamoDB table.
func (ddb *DynamoDB[T]) PutItem(item T) error {
	marshalledItem, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	_, err = ddb.client.PutItem(
		context.TODO(),
		&dynamodb.PutItemInput{TableName: aws.String(ddb.table), Item: marshalledItem},
	)
	if err != nil {
		return err
	}

	return nil
}

// Perform an update to an item
func (ddb *DynamoDB[T]) UpdateItem(key string, findValue string, item T) error {
	// Create a new expression builder
	updater := expression.Set(expression.Name("Modified"), expression.Value(time.Now().Unix()))

	// Iterate over the fields of the struct and add them to the expression builder
	t := reflect.TypeOf(item)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Ignore unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Check if the value is empty
		if IsEmptyValue(reflect.ValueOf(item).Field(i)) {
			continue
		}

		// Maybe i do not fully understand how reflect works so I get it again
		value := reflect.ValueOf(item).Field(i).Interface()

		// Add the field to the expression builder
		updater = updater.Set(expression.Name(field.Name), expression.Value(value))
	}

	// Build the expression
	expr, err := expression.NewBuilder().WithUpdate(updater).Build()
	if err != nil {
		return err
	}

	// Marshal the find value
	marshalledFindValue, err := attributevalue.Marshal(findValue)
	if err != nil {
		panic(err)
	}

	_, err = ddb.client.UpdateItem(
		context.TODO(),
		&dynamodb.UpdateItemInput{
			TableName:                 aws.String(ddb.table),
			Key:                       map[string]types.AttributeValue{key: marshalledFindValue},
			UpdateExpression:          expr.Update(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func IsEmptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Ptr:
		return value.IsNil()
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if !IsEmptyValue(value.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if value.Len() > 0 {
			return false
		}
		return true
	default:
		return true
	}
}
