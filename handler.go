/*
Author : Alexis Mimran

Lambda function which untar .ova files pushed on an S3 bucket

To automate the execution of this function, add a trigger on the lambda

Env parameters :
- DESTINATION_S3_BUCKET (if unset, destination bucket = source bucket)
*/

package main

import (
	"archive/tar"
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func HandleRequest(ctx context.Context, event events.S3Event) {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create S3 service client
	svc := s3.New(sess)

	// For each record
	for _, record := range event.Records {
		sourceBucket := record.S3.Bucket.Name
		log.Println("Source bucket : ", sourceBucket)
		sourceKey := record.S3.Object.Key
		log.Println("Source key : ", sourceKey)
		var destinationBucket string
		if os.Getenv("DESTINATION_S3_BUCKET") == "" {
			destinationBucket = sourceBucket
		} else {
			log.Println("Environment variable DESTINATION_S3_BUCKET is not set")
			destinationBucket = os.Getenv("DESTINATION_S3_BUCKET")
		}
		log.Println("Destination bucket : ", destinationBucket)
		rawObject, err := svc.GetObject(
			&s3.GetObjectInput{
				Bucket: &sourceBucket,
				Key:    &sourceKey,
			})

		if err != nil {
			log.Fatalf("Failed to read file from S3: %s", err)
		}
		tarReader := tar.NewReader(rawObject.Body)
		folderName := strings.TrimSuffix(sourceKey, ".ova")
		// If sourceKey doesn't end by .ova
		if sourceKey == folderName {
			log.Fatalf("The source file is not an OVA (file extension is not .ova)")
		}
		uploader := s3manager.NewUploader(sess)
		for true {
			// For each file in the tarball
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
			}
			if err != nil {
				log.Fatalf("Failed to open file: %s", err)
			}
			switch header.Typeflag {
			case tar.TypeReg:
				path := folderName + "/" + header.Name
				log.Println("Creating file:", header.Name)
				result, err := uploader.Upload(&s3manager.UploadInput{
					Bucket: &destinationBucket,
					Key:    aws.String(path),
					Body:   tarReader,
				})
				if err != nil {
					log.Fatalln("Failed to upload", err)
				}
				log.Println("Uploaded", header.Name, result.Location)

			default:
				log.Fatalf(
					"ExtractTarGz: uknown type: %s in %s",
					header.Typeflag,
					header.Name)
			}
		}
	}
}

func main() {
	lambda.Start(HandleRequest)
}
