package sync

import (
	"fmt"
	"log/slog"
	"net/url"
	"provider/football_org/api"
	"time"
)

const DAYS_TO_FETCH = 3

func FetchAPI(competitionID uint, mostRecentTimestamp *time.Time) (api.MatchesResponse, error) {
	apiPath, err := buildAPIPath(competitionID, mostRecentTimestamp)
	if err != nil {
		return api.MatchesResponse{}, err
	}

	matchesResponse, err := api.GetList(apiPath)
	if err != nil {
		slog.Error(err.Error())
		return api.MatchesResponse{}, err
	}

	return matchesResponse, nil
}

func buildAPIPath(competitionID uint, mostRecentTimestamp *time.Time) (string, error) {
	path := fmt.Sprintf("/competitions/%d/matches", competitionID)

	from, to := calculateDateRange(mostRecentTimestamp)
	params := url.Values{}
	params.Add("dateFrom", from.Format("2006-01-02"))
	params.Add("dateTo", to.Format("2006-01-02"))

	queryString := params.Encode()
	return path + "?" + queryString, nil
}

func calculateDateRange(mostRecentTimestamp *time.Time) (time.Time, time.Time) {
	var from time.Time
	if mostRecentTimestamp == nil {
		from = time.Now()
	} else {
		from = mostRecentTimestamp.Add(24 * time.Hour)
	}

	to := from.Add(DAYS_TO_FETCH * 24 * time.Hour)
	return from, to
}
