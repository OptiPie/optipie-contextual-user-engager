package dbmodels

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

type User struct {
	UserName             string    `dynamodbav:"user_name"`
	Created              time.Time `dynamodbav:"created"`
	IsReplied            bool      `dynamodbav:"is_replied"`
	RepliedTweetCount    int       `dynamodbav:"replied_tweet_count"`
	LastRepliedTweetTime time.Time `dynamodbav:"last_replied_tweet_time"`
}

type UpdateUserArgs struct {
	IsReplied            bool
	RepliedTweetCount    int
	LastRepliedTweetTime time.Time
}

func (u User) GetPrimaryKey() (map[string]types.AttributeValue, error) {
	userName, err := attributevalue.Marshal(u.UserName)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"user_name": userName}, nil
}
