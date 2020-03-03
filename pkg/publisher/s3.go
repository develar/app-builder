package publisher

import (
	"context"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
)

type ObjectOptions struct {
	file *string

	endpoint *string
	region   *string
	bucket   *string
	key      *string

	acl          *string
	storageClass *string
	encryption   *string

	accessKey *string
	secretKey *string
}

func ConfigurePublishToS3Command(app *kingpin.Application) {
	command := app.Command("publish-s3", "Publish to S3")
	options := ObjectOptions{
		file: command.Flag("file", "").Required().String(),

		region:   command.Flag("region", "").String(),
		bucket:   command.Flag("bucket", "").Required().String(),
		key:      command.Flag("key", "").Required().String(),
		endpoint: command.Flag("endpoint", "").String(),

		acl:          command.Flag("acl", "").String(),
		storageClass: command.Flag("storageClass", "").String(),
		encryption:   command.Flag("encryption", "").String(),

		accessKey: command.Flag("accessKey", "").String(),
		secretKey: command.Flag("secretKey", "").String(),
	}

	command.Action(func(context *kingpin.ParseContext) error {
		err := upload(&options)
		if err != nil {
			return err
		}
		return nil
	})

	configureResolveBucketLocationCommand(app)
}

func configureResolveBucketLocationCommand(app *kingpin.Application) {
	command := app.Command("get-bucket-location", "")
	bucket := command.Flag("bucket", "").Required().String()
	command.Action(func(parseContext *kingpin.ParseContext) error {
		requestContext, _ := util.CreateContextWithTimeout(30*time.Second)
		result, err := getBucketRegion(aws.NewConfig(), *bucket, requestContext, createHttpClient())
		if err != nil {
			return err
		}

		_, err = os.Stdout.WriteString(result)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
}

func getBucketRegion(awsConfig *aws.Config, bucket string, context context.Context, httpClient *http.Client) (string, error) {
	awsSession, err := session.NewSession(awsConfig, &aws.Config{
		// any region required
		Region:     aws.String("us-east-1"),
		HTTPClient: httpClient,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	result, err := s3manager.GetBucketRegion(context, awsSession, bucket, "")
	if err != nil {
		if awsError, ok := err.(awserr.Error); ok && awsError.Code() == "NotFound" {
			return "", errors.Errorf("unable to find bucket %s's region not found", bucket)
		}
		return "", errors.WithStack(err)
	}
	return result, nil
}

func upload(options *ObjectOptions) error {
	publishContext, _ := util.CreateContext()

	httpClient := createHttpClient()

	awsConfig := &aws.Config{
		HTTPClient: httpClient,
	}
	if *options.endpoint != "" {
		awsConfig.Endpoint = options.endpoint
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	//awsConfig.WithLogLevel(aws.LogDebugWithHTTPBody)

	if *options.accessKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(*options.accessKey, *options.secretKey, "")
	}

	switch {
	case *options.region != "":
		awsConfig.Region = options.region
	case *options.endpoint != "":
		awsConfig.Region = aws.String("us-east-1")
	default:
		// AWS SDK for Go requires region
		region, err := getBucketRegion(awsConfig, *options.bucket, publishContext, httpClient)
		if err != nil {
			return errors.WithStack(err)
		}
		awsConfig.Region = &region
	}

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		return errors.WithStack(err)
	}

	uploader := s3manager.NewUploader(awsSession)

	file, err := os.Open(*options.file)
	defer util.Close(file)
	if err != nil {
		return errors.WithStack(err)
	}

	uploadInput := s3manager.UploadInput{
		Bucket:      options.bucket,
		Key:         options.key,
		ContentType: aws.String(getMimeType(*options.key)),
		Body:        file,
	}
	if *options.acl != "" {
		uploadInput.ACL = options.acl
	}
	if *options.storageClass != "" {
		uploadInput.StorageClass = options.storageClass
	}
	if *options.encryption != "" {
		uploadInput.ServerSideEncryption = options.encryption
	}

	_, err = uploader.UploadWithContext(publishContext, &uploadInput)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func createHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: util.ProxyFromEnvironmentAndNpm,
		},
	}
}

func getMimeType(key string) string {
	if strings.HasSuffix(key, ".AppImage") {
		return "application/vnd.appimage"
	}
	if strings.HasSuffix(key, ".exe") {
		return "application/octet-stream"
	}
	if strings.HasSuffix(key, ".zip") {
		return "application/zip"
	}
	if strings.HasSuffix(key, ".blockmap") {
		return "application/gzip"
	}
	if strings.HasSuffix(key, ".snap") {
		return "application/vnd.snap"
	}
	if strings.HasSuffix(key, ".dmg") {
		//noinspection SpellCheckingInspection
		return "application/x-apple-diskimage"
	}

	ext := path.Ext(key)
	if ext != "" {
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			return mimeType
		}
	}
	return "application/octet-stream"
}
