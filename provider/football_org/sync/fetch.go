package sync

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"provider/football_org/api"
	"time"
)

const DAYS_TO_FETCH = 3

func FetchAPI(ctx context.Context, competitionID uint, mostRecentTimestamp *time.Time) (api.MatchesResponse, error) {
	apiPath := buildAPIPath(competitionID, mostRecentTimestamp)
	matchesResponse, err := api.GetList(ctx, apiPath)
	if err != nil {
		slog.Error(err.Error())
		return api.MatchesResponse{}, err
	}

	return matchesResponse, nil
}

func buildAPIPath(competitionID uint, mostRecentTimestamp *time.Time) (string) {
	path := fmt.Sprintf("/competitions/%d/matches", competitionID)

	from, to := calculateDateRange(mostRecentTimestamp)
	params := url.Values{}
	params.Add("dateFrom", from.Format("2006-01-02"))
	params.Add("dateTo", to.Format("2006-01-02"))

	queryString := params.Encode()
	return path + "?" + queryString
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
