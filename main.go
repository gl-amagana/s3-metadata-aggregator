package main

import "log"

func main() {
	// Get existing spreadsheet
	spreadsheet := setupSheet(SpreadsheetId)

	// Collect data
	bucketMetadataCollection := getAllBucketMetadata()

	// Set spreadsheet data
	populateSpreadsheet(&spreadsheet, bucketMetadataCollection)
	log.Printf("Spreadsheet URL: %s\t", spreadsheet.spreadsheetUrl)
}
