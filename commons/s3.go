package commons

import (
	"github.com/deis/go-boot-proposal/logger"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
)

// ConnectToS3Store created a connection to a S3 compatible store (Ceph)
func ConnectToS3Store(accessKey string, secretKey string, host string, region string) *s3.S3 {
	logger.Log.Debugf("connecting to S3 data store located in %v", host)
	auth := aws.Auth{AccessKey: accessKey, SecretKey: secretKey}
	// if region == nil{
	//   region = "deis-region-1"
	// }
	return s3.New(auth, aws.Region{Name: "deis-region-1", S3Endpoint: host})
}
