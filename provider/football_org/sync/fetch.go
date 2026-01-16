package sync

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"provider/football_org/api"
	"time"
)

func FetchAPI(ctx context.Context, competitionID uint, from time.Time, to time.Time) (api.MatchesResponse, error) {
	apiPath := buildAPIPath(competitionID, from, to)
	matchesResponse, err := api.GetList(ctx, apiPath)
	if err != nil {
		slog.Error(err.Error())
		return api.MatchesResponse{}, err
	}

	return matchesResponse, nil
}

func buildAPIPath(competitionID uint, from time.Time, to time.Time) string {
	path := fmt.Sprintf("/competitions/%d/matches", competitionID)

	params := url.Values{}
	params.Add("dateFrom", from.Format("2006-01-02"))
	params.Add("dateTo", to.Format("2006-01-02"))

	queryString := params.Encode()
	return path + "?" + queryString
}
