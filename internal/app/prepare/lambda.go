package prepare

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

func Lambda(awsCfg aws.Config) *lambda.Client {
	return lambda.NewFromConfig(awsCfg)
}
