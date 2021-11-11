package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/aws/aws-sdk-go/service/s3"
	"sync"
)

/* TODO BLOCK ---

Notes from Alex:
	* Concurrency. Use Mutex or whatever tool needed to call all the profiles, grab data, populate structs.
		* Once that is done (from threads), go back to main and then call populateData() with new data struct.
	* For now, no worries on cron.
	* Spreadsheet ID is static, we should manually create once, then append sheets w/ name as date.now() every run.
*/

func main() {
	// Get existing spreadsheet

	//profiles := getAllProfiles()
	profiles := []string{"dev", "staging"}

	metadataGetters := []*MetaDataGetter{}

	for _, profile := range profiles {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           profile,
		}))
		metadataGetters = append(metadataGetters, &MetaDataGetter{
			macie: macie2.New(sess),
			s3:    s3.New(sess),
		})
	}

	results := &BucketMetaDataCollection{}
	wg := sync.WaitGroup{}

	for _, getter := range metadataGetters {
		wg.Add(1)
		co := func(g *MetaDataGetter) {
			bucketResults, err := g.describeAllBuckets()
			panic(err)
			results.Insert(bucketResults...)
			wg.Done()
		}

		go co(getter)
	}

	wg.Wait()

	// Set section

	//spreadsheet, _ := getSpreadsheet(SpreadsheetId)
	//setupSpreadsheet(spreadsheet.SpreadsheetId)
	//populateSpreadsheet(spreadsheet.SpreadsheetId)
	//log.Printf("Spreadsheet URL: %s\t\n", spreadsheet.SpreadsheetUrl)
}
