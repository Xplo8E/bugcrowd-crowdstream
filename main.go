package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// CustomTime represents a custom time format with JSON unmarshalling support.
type CustomTime struct {
	time.Time
}

// Dateonly represents a date-only format with JSON unmarshalling support.
type Dateonly struct {
	time.Time
}

// Response represents the JSON response structure from BugCrowd API.
type Response struct {
	Results []struct {
		Name            string     `json:"program_name"`
		Code            string     `json:"program_code"`
		Asset           string     `json:"target"`
		Severity        int        `json:"priority"`
		Reported        CustomTime `json:"created_at"`
		Accepted        Dateonly   `json:"accepted_at"`
		Points          int        `json:"points"`
		Bounty          string     `json:"amount"`
		Hacker          string     `json:"researcher_username"`
		Isprivate       bool       `json:"visibility_public"`
		Submission_text string     `json:"submission_state_text"`
	} `json:"results"`
}

const reportedDateLayout = time.RFC3339
const acceptedLayout = "2 Jan 2006"

// UnmarshalJSON implements JSON unmarshalling for CustomTime.
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(reportedDateLayout, s)
	return
}

// UnmarshalJSON implements JSON unmarshalling for Dateonly.
func (ct *Dateonly) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(acceptedLayout, s)
	return
}
// https://bugcrowd.com/crowdstream.json?page=1&filter_by=accepted%2Cdisclosures
func main() {
	page := 1
	// Encode the query parameters properly using url.Values
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("filter_by", "accepted,disclosures")

	// Use the encoded URL in the HTTP request
	crowdstream := fmt.Sprintf("https://bugcrowd.com/crowdstream.json?%s", params.Encode())

	fmt.Println("URL : ", crowdstream)

	resp, err := http.Get(crowdstream)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	fmt.Println("JSON Response:")
	fmt.Println(string(body))

	var jsonbody Response

	if err := json.Unmarshal(body, &jsonbody); err != nil {
		log.Println("Error in JSON Unmarshalling:", err)
		log.Println("JSON Response:", string(body))
		log.Fatalln(err)
	}

	// Create a folder for the current year
	year := time.Now().Year()
	yearFolder := fmt.Sprintf("%d", year)

	// Create a folder for the current month inside the year folder
	month := time.Now().Month().String()
	monthFolder := filepath.Join(yearFolder, month)

	// Ensure that the year and month folders exist
	foldercreate(yearFolder)
	foldercreate(monthFolder)

		// Create the new date-README.md file in the month folder
	dateReadmeFile, err := os.Create(filepath.Join(monthFolder, fmt.Sprintf("%d-%02d-%02d-README.md", year, time.Now().Month(), time.Now().Day())))
		if err != nil {
			log.Println("Error in file creation", err)
		}
	
		// Write content to the date-README.md file
		writeContent(dateReadmeFile, jsonbody)
	
			// Create the new README.md file in the main directory
	readmeFile, err := os.Create("README.md")
	if err != nil {
		log.Println("Error in file creation", err)
	}

	// Write content to the README.md file
	writeContent(readmeFile, jsonbody)

	// Move existing files to corresponding year/month folders
	//moveFilesToYearMonthFolders("archive")

	fmt.Println("Program Completed.")

}

// writeContent writes the program details to the provided file.
func writeContent(file *os.File, jsonbody Response) {
	fmt.Fprintln(file, "[![schedule run](https://github.com/Linuxinet/bugcrowd-crowdstream/actions/workflows/actions.yml/badge.svg?branch=master)](https://github.com/Linuxinet/bugcrowd-crowdstream/actions/workflows/actions.yml)")

	fmt.Fprintln(file, "## BugCrowd Crowdstream | Date: ", time.Now().Format("2006-January-02 15:04:05"))

	for i, p := range jsonbody.Results {
		// Write program details to the file
		fmt.Fprintf(file, "### %d. Program Details : \n\n**Name:** %s \n\n **Link:** <https://bugcrowd.com/%s> \n\n **Severity:** P%d \n\n **Hacker:** %s \n\n **Points:** %d \n\n **Target:** ` %s` \n\n **Reported:** %s \n\n **Accepted:** %s \n\n **%s** \n\n", i+1, p.Name, p.Code, p.Severity, p.Hacker, p.Points, p.Asset, p.Reported, p.Accepted, p.Submission_text)
	}

	fmt.Fprintln(file, "## End of Crowdstream for", time.Now().Format("2006-January-02 15:04:05"))

	file.Close()
}

// foldercreate creates the specified folder if it does not exist.
func foldercreate(folderPath string) {
	err := os.MkdirAll(folderPath, 0700)
	if err != nil {
		log.Println("Error in Creating Directory: ", folderPath)
		log.Fatalln(err)
	}
}

// moveFilesToYearMonthFolders moves files from the source directory to corresponding year/month folders.
func moveFilesToYearMonthFolders(sourceDir string) {
	// List all files in the source directory
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatalln("error in listing archive directory")
	}

	// Iterate through each file and move it to the corresponding year/month folder
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()

			// Extract the date from the filename (assuming filename is in the format DD-MM-YYYY-README.md)
			parts := strings.Split(fileName, "-")
			if len(parts) == 4 {
				// day, _ := strconv.Atoi(parts[0])
				month, _ := strconv.Atoi(parts[1])
				year, _ := strconv.Atoi(parts[2])

				// Create a folder for the current year
				yearFolder := fmt.Sprintf("%d", year)

				// Create a folder for the current month inside the year folder
				monthFolder := filepath.Join(yearFolder, time.Month(month).String())

				// Ensure that the year and month folders exist
				foldercreate(yearFolder)
				foldercreate(monthFolder)

				// Move the file to the year/month folder
				sourceFilePath := filepath.Join(sourceDir, fileName)
				destFilePath := filepath.Join(monthFolder, fileName)

				err := os.Rename(sourceFilePath, destFilePath)
				if err != nil {
					log.Printf("Error moving file %s to %s: %s\n", sourceFilePath, destFilePath, err)
				}
			}
		}
	}
}
