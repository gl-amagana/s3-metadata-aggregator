package main

import "log"

func main() {
	// Get existing spreadsheet
	spreadsheet, _ := getSpreadsheet(SpreadsheetId)
	sheetTitle := setupSheet(spreadsheet.SpreadsheetId)

	// Collect data
	bucketMetadataCollection := getAllBucketMetadata()

	// Set spreadsheet data
	populateSpreadsheet(spreadsheet.SpreadsheetId, sheetTitle, bucketMetadataCollection)
	log.Printf("Spreadsheet URL: %s\t\n", spreadsheet.SpreadsheetUrl)
}
