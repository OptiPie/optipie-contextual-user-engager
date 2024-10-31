package prepare

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func Dynamodb(awsCfg aws.Config) *dynamodb.Client {
	svc := dynamodb.NewFromConfig(awsCfg)
	return svc
}
