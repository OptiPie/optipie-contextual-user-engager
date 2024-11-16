package prepare

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func Ec2(awsCfg aws.Config) *ec2.Client {
	return ec2.NewFromConfig(awsCfg)
}
