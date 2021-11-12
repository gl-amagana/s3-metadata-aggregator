package main

import "log"

/* TODO BLOCK ---

Notes from Alex:
	* Concurrency. Use Mutex or whatever tool needed to call all the profiles, grab data, populate structs.
		* Once that is done (from threads), go back to main and then call populateData() with new data struct.
	* For now, no worries on cron.
	* Spreadsheet ID is static, we should manually create once, then append sheets w/ name as date.now() every run.
*/

func main() {
	// Get existing spreadsheet
	spreadsheet, _ := getSpreadsheet(SpreadsheetId)
	sheetTitle := setupSpreadsheet(spreadsheet.SpreadsheetId)

	// Collect data
	bucketMetadataCollection := getAllBucketMetadata()

	// Set spreadsheet data
	populateSpreadsheet(spreadsheet.SpreadsheetId, sheetTitle, bucketMetadataCollection)
	log.Printf("Spreadsheet URL: %s\t\n", spreadsheet.SpreadsheetUrl)
}
