package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"os"
)

const API_ENDPOINT = "https://api.football-data.org/v4"

func sync_football_org() {
	// LaLiga
	base, err := url.Parse(API_ENDPOINT + "/competitions/2014/matches")
	if err != nil {
		printError(fmt.Sprintf("failed to parse base URL: %v", err))
		return
	}

	// TODO: calculations (make sure this algorithm is correct)
	// from: here we should go to the db to get the most recent match stored.
	//  If we don't have any, we start from now.
	//  If from is already one week from now, we stop execution.
	// to: we add 1 week to from.
	from := time.Now()

	// Calculate the time 7 days (1 week) from now
	to := from.Add(7 * 24 * time.Hour)

	// 2. Create a new Values object for query parameters.
	// This is the cleanest way to handle query parameters as it automatically handles URL encoding.
	params := url.Values{}
	// Add your query parameters here
	params.Add("dateFrom", from.Format("2006-01-02"))
	params.Add("dateTo", to.Format("2006-01-02"))
	//params.Add("status", "FINISHED")

	// 3. Encode the parameters and append them to the base URL.
	base.RawQuery = params.Encode()
	finalURL := base.String()

	fmt.Printf("%s\n[INFO] Sending GET request to: %s%s\n", ColorYellow, finalURL, ColorReset)

	// 4. Create a new HTTP request with custom headers
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		printError(fmt.Sprintf("failed to create request: %v", err))
		return
	}

	// Add the X-Auth-Token header
	req.Header.Set("X-Auth-Token", os.Getenv("FOOTBALL_ORG_API_KEY"))

	// 5. Execute the GET request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		printError(fmt.Sprintf("failed to get matches: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printError(fmt.Sprintf("failed to read response body: %v", err))
		return
	}

	fmt.Printf("%s\n[INFO] Response body: %s%s\n", ColorGreen, string(body), ColorReset)
}
