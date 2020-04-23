package screenshot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
)

// BmcScreenshoter represents BMC providers with the minimal set of methods
type BmcScreenshoter interface {
	Screenshot() ([]byte, string, error)
	HardwareType() string
}

// Upload screenshots to a S3 bucket
// TODO: this is a simple poc with s3, we need more love to make it better.
func upload(payload []byte, fileName string) (url string, err error) {
	bucket := aws.String(viper.GetString("s3.bucket"))
	key := aws.String(fmt.Sprintf("%s/%s", viper.GetString("s3.folder"), fileName))

	fmt.Println(*key)
	// Configure to use Minio Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(viper.GetString("s3.access_key_id"), viper.GetString("s3.secret_access_key"), ""),
		Endpoint:         aws.String(viper.GetString("s3.endpoint")),
		Region:           aws.String(viper.GetString("s3.region")),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return url, err
	}
	s3Client := s3.New(newSession)
	cparams := &s3.CreateBucketInput{
		Bucket: bucket, // Required
	}

	// Create a new bucket using the CreateBucket call.
	_, err = s3Client.CreateBucket(cparams)
	if err != nil {
		return url, err
	}

	// Upload a new object "testobject" with the string "Hello World!" to our "newbucket".
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Body:          bytes.NewReader(payload),
		ContentType:   aws.String(http.DetectContentType(payload)),
		Bucket:        bucket,
		ContentLength: aws.Int64(int64(binary.Size(payload))),
		Key:           key,
		ACL:           aws.String(viper.GetString("s3.acl")),
	})
	if err != nil {
		return url, fmt.Errorf("failed to upload data to %s/%s: %s", *bucket, *key, err.Error())
	}

	return fmt.Sprintf("%s/%s%s", viper.GetString("s3.endpoint"), *bucket, *key), err
}

func takeScreenShot(bmc BmcScreenshoter, host string) (payload []byte, fileName string, err error) {
	payload, extension, err := bmc.Screenshot()
	if err != nil {
		return payload, fileName, err
	}

	fileName = fmt.Sprintf(
		"%s-%s-%d.%s",
		host,
		bmc.HardwareType(),
		time.Now().Unix(),
		extension,
	)

	return payload, fileName, err
}

// S3 takes a screenshot and upload to s3
func S3(bmc BmcScreenshoter, host string) (fileURL string, status bool, err error) {
	payload, fileName, err := takeScreenShot(bmc, host)
	if err != nil {
		return fileURL, status, err
	}

	fileURL, err = upload(payload, fileName)
	if err == nil {
		status = true
	}

	return fileURL, status, err
}

// Local takes screenshot and store it locally on the server
func Local(bmc BmcScreenshoter, host string) (fileURL string, status bool, err error) {
	payload, fileName, err := takeScreenShot(bmc, host)
	if err != nil {
		return fileURL, status, err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", viper.GetString("screenshot_storage"), fileName), payload, 0644)
	if err != nil {
		return fileURL, false, err
	}

	return fmt.Sprintf("/screenshot/%s", fileName), true, err
}
