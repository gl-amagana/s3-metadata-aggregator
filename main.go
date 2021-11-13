package main

import "log"

func main() {
	// Get existing spreadsheet
	sheetTitle := setupSheet(SpreadsheetId)

	// Collect data
	bucketMetadataCollection := getAllBucketMetadata()

	// Set spreadsheet data
	populateSpreadsheet(spreadsheet.SpreadsheetId, sheetTitle, bucketMetadataCollection)
	log.Printf("Spreadsheet URL: %s\t\n", spreadsheet.SpreadsheetUrl)
}
