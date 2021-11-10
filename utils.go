package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"log"
)

type Profiles struct {
	Profiles []string `json:"profiles"`
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
