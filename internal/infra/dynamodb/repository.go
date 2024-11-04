package dynamodbrepo

import (
	"context"
	"fmt"
	dbmodels "github.com/OptiPie/optipie-contextual-user-engager/internal/infra/dynamodb/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

const (
	userPrimaryKey = "user_name"

	// dynamodb condition expressions
	attributeNotExists = "attribute_not_exists"
	attributeExists    = "attribute_exists"
)

type Client struct {
	client         *dynamodb.Client
	usersTableName string
}

func NewRepository(client *dynamodb.Client, usersTableName string) *Client {
	return &Client{
		client:         client,
		usersTableName: usersTableName,
	}
}

// CreateUser creates initial list of users, it will be only used to populate table once
func (c *Client) CreateUser(ctx context.Context, userName string) error {
	user := dbmodels.User{
		UserName:             userName,
		Created:              time.Now(),
		IsReplied:            false,
		RepliedTweetCount:    0,
		LastRepliedTweetTime: time.Now(),
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("marshall_map error: %v", err)
	}

	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.usersTableName), Item: item,
	})
	if err != nil {
		return fmt.Errorf("put_item error: %v", err)
	}

	return nil
}

// GetUserNamesToReply to retrieve list of user names that has not been replied yet
func (c *Client) GetUserNamesToReply(ctx context.Context) ([]string, error) {
	filter := expression.Name("is_replied").Equal(expression.Value(false))
	projection := expression.NamesList(expression.Name("user_name"))

	expr, err := expression.NewBuilder().WithFilter(filter).WithProjection(projection).Build()
	if err != nil {
		return nil, fmt.Errorf("expression newBuilder error: %v", err)
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(c.usersTableName),
	}

	result, err := c.client.Scan(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("client scan error: %v", err)
	}
	userNames := make([]string, result.Count)

	for i, item := range result.Items {
		var user dbmodels.User

		err = attributevalue.UnmarshalMap(item, &user)
		if err != nil {
			return nil, fmt.Errorf("unmarshalMap error: %v", err)
		}
		userNames[i] = user.UserName
	}
	return userNames, nil
}

// GetUsers to retrieve list of all users
func (c *Client) GetUsers(ctx context.Context) ([]dbmodels.User, error) {
	params := &dynamodb.ScanInput{
		TableName: aws.String(c.usersTableName),
	}

	result, err := c.client.Scan(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("client scan error: %v", err)
	}

	users := make([]dbmodels.User, result.Count)

	for i, item := range result.Items {
		var user dbmodels.User

		err = attributevalue.UnmarshalMap(item, &user)
		if err != nil {
			return nil, fmt.Errorf("unmarshalMap error: %v", err)
		}
		users[i] = user
	}
	return users, nil
}

func (c *Client) UpdateUser(ctx context.Context, userName string, args dbmodels.UpdateUserArgs) error {
	update := expression.Set(expression.Name("is_replied"), expression.Value(args.IsReplied))
	update.Set(expression.Name("last_replied_tweet_time"), expression.Value(args.LastRepliedTweetTime))
	if args.RepliedTweetCount != 0 {
		update.Set(expression.Name("replied_tweet_count"), expression.Value(args.RepliedTweetCount))
	} else {
		update.Add(expression.Name("replied_tweet_count"), expression.Value(1))
	}

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return fmt.Errorf("expression_builder error: %v", err)
	}

	user := dbmodels.User{UserName: userName}
	userPk, err := user.GetPrimaryKey()
	if err != nil {
		return fmt.Errorf("get_primary_key error: %v", err)
	}

	conditionExpression := fmt.Sprintf("%v(%v)", attributeExists, userPrimaryKey)

	_, err = c.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(c.usersTableName),
		Key:                       userPk,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       aws.String(conditionExpression),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueNone,
	})

	if err != nil {
		return fmt.Errorf("update_item error: %v", err)
	}

	return nil
}
