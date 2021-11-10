package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// getClient - Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// tokenFromFile - Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken - Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	_ = json.NewEncoder(f).Encode(token)
}

// getTokenFromWeb - Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// sheetService - Creates a Google Sheet client
func sheetService() *sheets.Service {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	return srv
}

// getSpreadsheet - Gets an existing Google spreadsheet
func getSpreadsheet(spreadsheetId string) (*sheets.Spreadsheet, error) {
	srv := sheetService()
	response, err := srv.Spreadsheets.Get(spreadsheetId).Do()
	if err != nil {
		log.Fatalf("Unable to find spreadsheet: %v\n", err)
		return nil, err
	}
	log.Println("Found existing spreadsheet")
	return response, nil
}

// createSpreadsheet - Creates a new Google Spreadsheet
func createSpreadsheet() sheets.Spreadsheet {
	srv := sheetService()
	currentTime := time.Now().Format("2006-02-01")
	spreadsheetTitle := "S3 Report - " + currentTime
	spreadsheet := sheets.Spreadsheet{Properties: &sheets.SpreadsheetProperties{Title: spreadsheetTitle}}

	response, err := srv.Spreadsheets.Create(&spreadsheet).Do()
	if err != nil {
		log.Fatalf("Unable to create spreadsheet: %v", err)
		return sheets.Spreadsheet{}
	}

	log.Println("Creating spreadsheet")
	return *response
}

// setupSpreadsheet - Adds header values to spreadsheet
func setupSpreadsheet(spreadsheetId string) {
	srv := sheetService()

	// Spreadsheet Headers
	valueRange := sheets.ValueRange{}
	headers := []interface{}{"Account ID", "Bucket Name", "Encryption", "isVersioned?", "No. of Objects Unencrypted", "isLoggingEnabled?"} // TODO: Maybe add account ID field?
	valueRange.Values = append(valueRange.Values, headers)

	_, err := srv.Spreadsheets.Values.Update(spreadsheetId, "A1", &valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		log.Fatalf("Unable to write to spreadsheet: %v", err)
	}
}

// populateSpreadsheet - Populates appropriate columns; must match headers
func populateSpreadsheet(spreadsheetId string) {
	callerIdentity, _ := getCallerIdentity()
	srv := sheetService()

	writeRangeData := "A2"
	sheetData := sheets.ValueRange{}

	buckets, _ := describeAllBuckets()

	log.Printf("Running for buckets in account: %s...\n", *callerIdentity.Account)
	for _, bucket := range buckets {
		dataInput := []interface{}{bucket.accountId, bucket.bucketName, bucket.encryptionType, bucket.isVersioned, bucket.objectsEncrypted, bucket.isLogging}
		sheetData.Values = append(sheetData.Values, dataInput)

		_, err := srv.Spreadsheets.Values.Update(spreadsheetId, writeRangeData, &sheetData).ValueInputOption("RAW").Do()
		if err != nil {
			log.Fatalf("Unable to write to spreadsheet: %v\n", err)
		}
		time.Sleep(10)
	}
	log.Println("Complete.")
}

func updateSheet(spreadsheetId string, req *sheets.Request) (*sheets.BatchUpdateSpreadsheetResponse, error) {
	srv := sheetService()

	update := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	response, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, &update).Do()
	if err != nil {
		log.Fatalf("Unable to update spreadsheet: %v", err)
		return nil, err
	}

	return response, nil
}
