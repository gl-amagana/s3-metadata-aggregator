package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/aws/aws-sdk-go/service/macie2/macie2iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts"
	"log"
	"sync"
)

type BucketMetadata struct {
	accountId        string
	bucketName       string
	encryptionType   string
	objectsEncrypted int64
	isVersioned      bool
	isLogging        bool
}

type BucketMetaDataCollection struct {
	lock  sync.Mutex
	items []*BucketMetadata
}

func (b *BucketMetaDataCollection) Insert(result ...*BucketMetadata) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.items = append(b.items, result...)
}

func (b *BucketMetaDataCollection) GetItems() []*BucketMetadata{
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.items
}

type MetaDataGetter struct {
	macie macie2iface.Macie2API
	s3 s3iface.S3API
}

//getAwsSession - Returns AWS Session
//func getAwsSession() *session.Session {
//	return session.Must(session.NewSessionWithOptions(session.Options{
//		SharedConfigState: session.SharedConfigEnable,
//		Profile:           string("staging"),
//	}))
//}

// getCallerIdentity - Returns metadata about account
//func getCallerIdentity() (*sts.GetCallerIdentityOutput, error) {
//	svc := sts.New(getAwsSession())
//	result, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
//	if err != nil {
//		return nil, err
//	}
//
//	return result, nil
//}

// describeAllBuckets - Returns all buckets' metadata
func (gen *MetaDataGetter) describeAllBuckets() ([]*BucketMetadata, error) {

	var bucketMetadataList []*BucketMetadata
	count := 0
	err := gen.macie.DescribeBucketsPages(&macie2.DescribeBucketsInput{MaxResults: aws.Int64(50)}, func(page *macie2.DescribeBucketsOutput, lastPage bool) bool {
		for _, bucket := range page.Buckets {
			count++
			bucketMetadataList = append(bucketMetadataList, &BucketMetadata{
				accountId:        *bucket.AccountId,
				bucketName:       *bucket.BucketName,
				encryptionType:   *bucket.ServerSideEncryption.Type,
				objectsEncrypted: *bucket.ObjectCountByEncryptionType.Unencrypted,
				isVersioned:      *bucket.Versioning,
				isLogging:        getBucketLoggingByBucket(*bucket.BucketName),
			})
		}
		return !lastPage
	})
	log.Printf("Total Buckets in %s: %d", bucketMetadataList[0].accountId, count)
	if err != nil {
		return nil, err
	}

	return bucketMetadataList, nil
}

// getBucketLoggingByBucket - Returns bool if logging is enabled for given bucket
func getBucketLoggingByBucket(bucketName string) bool {
	result, err := getBucketLogging(bucketName)
	if err != nil {
		log.Printf("Got an error grabbing bucket ownership metadata: %v", err)
	}

	// go has no ternary operator... :cry:
	isLogging := false
	if result.LoggingEnabled != nil {
		isLogging = true
	}
	return isLogging
}

// getBucketLogging - Returns bucket logging data via s3
func getBucketLogging(bucket string) (*s3.GetBucketLoggingOutput, error) {
	svc := s3.New(getAwsSession())
	result, err := svc.GetBucketLogging(&s3.GetBucketLoggingInput{Bucket: &bucket})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Ref: https://stackoverflow.com/questions/52936693/aws-s3-bucket-encryption-bucket-property-setting-vs-bucket-policy-setting
