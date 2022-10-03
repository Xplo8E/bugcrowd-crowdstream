package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type CustomTime struct {
	time.Time
}

type Dateonly struct {
	time.Time
}
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

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(reportedDateLayout, s)
	return
}

func (ct *Dateonly) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(acceptedLayout, s)
	return
}

func main() {
	page := 1
	crowdstream := fmt.Sprintf("https://bugcrowd.com/crowdstream.json?page=%d&filter_by=all", page)

	resp, err := http.Get(crowdstream)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	var jsonbody Response

	if err := json.Unmarshal(body, &jsonbody); err != nil {
		log.Fatalln(err)
	}

	// folder checking
	archive := "archive"
	if _, err := os.Stat(archive); os.IsNotExist(err) {
		foldercreate(archive)
	} else {

		files, err := os.ReadDir(archive)
		if err != nil {
			log.Fatalln("error in listing archive directory")
		}
		if len(files) >= 100 {
			fmt.Printf("%s folder contains 100+ files\n", archive)
		}

	}

	// end folder checking

	// check file exist and rename to old name
	path := "README.md"
	old_name := fmt.Sprintf("archive/%v-README.md", time.Now().AddDate(0, 0, -1).Format("02-01-2006"))
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		fmt.Println("README.md file is not exists")
	} else {
		err := os.Rename(path, old_name)
		if err != nil {
			log.Fatalln("error in renaming filename", err)
		}
	}
	// end file checking

	readmefile, err := os.Create("README.md")
	if err != nil {
		log.Println("Error in file creation", err)
	}

	fmt.Println("Writing To ", readmefile.Name())

	fmt.Fprintln(readmefile, "[![schedule run](https://github.com/Linuxinet/bugcrowd-crowdstream/actions/workflows/actions.yml/badge.svg?branch=master)](https://github.com/Linuxinet/bugcrowd-crowdstream/actions/workflows/actions.yml)")

	fmt.Fprintln(readmefile, "## BugCrowd Crowdstream | Date: ", time.Now().Format("2006-January-02 15:04:05"))

	fmt.Fprintln(readmefile, "                            ")
	for i, p := range jsonbody.Results {

		fmt.Fprintf(readmefile, "### %d. Program Details : \n\n**Name:** %s \n\n **Link:** <https://bugcrowd.com/%s> \n\n **Severity:** P%d \n\n **Hacker:** %s \n\n **Points:** %d \n\n **Target:** ` %s` \n\n **Reported:** %s \n\n **Accepted:** %s \n\n **%s** \n\n", i+1, p.Name, p.Code, p.Severity, p.Hacker, p.Points, p.Asset, p.Reported, p.Accepted, p.Submission_text)

		//		fmt.Printf("%d. Program Details : \n\tName: %s \n\tCode: %s \n\tSeverity: P%d \n\tHacker: %s \n\tPoints: %d \n\tTarget: %s \n\tReported : %s \n\tAccepted: %s \n\t%s \n\n", i+1, p.Code, p.Severity, p.Hacker, p.Points, p.Asset, p.Reported, p.Accepted, p.Submission_text)

		/*		_, err := io.WriteString(readmefile, contentData)
				if err != nil {
					log.Fatalln("error in data write to file", err)
				}*/

		/*
			fmt.Printf("%d. Program Details : \n", i+1)
			fmt.Println("                            ")
			fmt.Printf("\tName: %s \n", p.Name)
			fmt.Printf("\tCode: %s \n", p.Code)
			fmt.Printf("\tSeverity: P%d \n", p.Severity)
			fmt.Printf("\tHacker: %s \n", p.Hacker)
			fmt.Printf("\tPoints: %d \n", p.Points)
			fmt.Printf("\tTarget: %s \n", p.Asset)
			fmt.Printf("\tReported : %s \n", p.Reported)
			fmt.Printf("\tAccepted: %s \n", p.Accepted)
			fmt.Printf("\tTarget: %s \n", p.Asset)
			fmt.Printf("\t%s \n", p.Submission_text)
			fmt.Println("                            ")
			fmt.Println("...............................")
			fmt.Println("                            ")
		*/
	}
	fmt.Fprintln(readmefile, "## End of Crowdstream for", time.Now().Format("2006-January-02 15:04:05"))
	fmt.Println("Program Completed.")
}

func foldercreate(archive string) {
	err := os.Mkdir(archive, 0700)
	if err != nil {
		log.Println("Error in Creating Directory: ", archive)
		log.Fatalln(err)
	}
}
