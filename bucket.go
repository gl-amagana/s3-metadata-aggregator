package main

import (
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/aws/aws-sdk-go/service/s3"
	"sync"
)

// NOTE: Just pass that BucketResultCollection struct in to whatever concurrent functions you want, will make sure your list appends are always thread safe.

type BucketMetadataInput struct {
	macie macie2.Macie2
	s3    s3.S3
}

type BucketMetadataOutput struct {
	bucketName       string
	encryptionType   string
	objectsEncrypted int
	isVersioned      bool
	isLogging        bool
}

type BucketResultCollection struct {
	lock  sync.Mutex
	items []BucketMetadataOutput
}

func (b *BucketMetadataInput) getData() chan BucketMetadataOutput {
	// TODO: This is where we call and concurrently run the scanners.
	return nil
}

func (b *BucketResultCollection) Insert(result ...BucketMetadataOutput) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.items = append(b.items, result...)
}

func (b *BucketResultCollection) GetItems() []BucketMetadataOutput {
	return b.items
}
