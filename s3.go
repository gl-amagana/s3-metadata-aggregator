package main

import (
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/aws/aws-sdk-go/service/s3"
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

func (b *BucketMetaDataCollection) GetItems() []*BucketMetadata {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.items
}

// describeAllBuckets - Returns all buckets' metadata
func (gen *AWSClients) describeAllBuckets() ([]*BucketMetadata, error) {
	var bucketMetadataList []*BucketMetadata
	count := 0
	err := gen.macie.DescribeBucketsPages(&macie2.DescribeBucketsInput{}, func(page *macie2.DescribeBucketsOutput, lastPage bool) bool {
		for _, bucket := range page.Buckets {
			count++
			bucketMetadataList = append(bucketMetadataList,
				&BucketMetadata{
					accountId:        *bucket.AccountId,
					bucketName:       *bucket.BucketName,
					encryptionType:   *bucket.ServerSideEncryption.Type,
					objectsEncrypted: *bucket.ObjectCountByEncryptionType.Unencrypted,
					isVersioned:      *bucket.Versioning,
					isLogging:        gen.getBucketLoggingByBucket(*bucket.BucketName),
				})
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return bucketMetadataList, nil
}

// getBucketLoggingByBucket - Returns bool if logging is enabled for given bucket
func (gen *AWSClients) getBucketLoggingByBucket(bucketName string) bool {
	result, err := gen.getBucketLogging(bucketName)
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
func (gen *AWSClients) getBucketLogging(bucket string) (*s3.GetBucketLoggingOutput, error) {
	result, err := gen.s3.GetBucketLogging(&s3.GetBucketLoggingInput{Bucket: &bucket})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getAllBucketMetadata() *BucketMetaDataCollection {
	log.Println("Collecting S3 bucket metadata...")
	awsClients := getAwsSessions()

	results := &BucketMetaDataCollection{}
	wg := sync.WaitGroup{}

	for _, client := range awsClients {
		wg.Add(1)

		co := func(client *AWSClients, results *BucketMetaDataCollection) {
			defer wg.Done()

			bucketResults, err := client.describeAllBuckets()
			if err != nil {
				log.Fatalf("Error describing all buckets: %v", err)
			}
			results.Insert(bucketResults...)
		}
		go co(client, results)
	}
	wg.Wait()

	return results
}

// Ref: https://stackoverflow.com/questions/52936693/aws-s3-bucket-encryption-bucket-property-setting-vs-bucket-policy-setting
