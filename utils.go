package main

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/aws/aws-sdk-go/service/macie2/macie2iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"io/ioutil"
	"os"

	"log"
)

type Profiles struct {
	Profiles []string `json:"profiles"`
}

type AWSClients struct {
	macie macie2iface.Macie2API
	s3    s3iface.S3API
	sts   stsiface.STSAPI
}

// getAllProfiles() will return all profiles from profiles.json
func getAllProfiles() []string {
	profiles := Profiles{}
	jsonFile, err := os.Open("profiles.json")
	input, _ := ioutil.ReadAll(jsonFile)
	e := json.Unmarshal(input, &profiles)
	if e != nil {
		log.Fatalf("Unable to unmarshal JSON file :%v", err)
	}

	// Close file
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	return profiles.Profiles
}

// getAwsSessions - Returns all necessary AWS clients needed
func getAwsSessions() []*AWSClients {
	// FIXME: macie not enabled for: kronaprod, kronanetworking, operations, secops
	allProfiles := getAllProfiles()

	var awsClients []*AWSClients
	var regions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2"}

	for _, profile := range allProfiles {
		for _, region := range regions {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
				Profile:           profile,
				Config:            aws.Config{Region: aws.String(region)},
			}))
			awsClients = append(awsClients, &AWSClients{
				macie: macie2.New(sess),
				s3:    s3.New(sess),
			})
		}
	}
	return awsClients
}
