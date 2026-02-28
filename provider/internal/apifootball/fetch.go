package apifootball

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"provider/internal/entity"
	"strconv"
	"time"
)

type apifootballEvent struct {
	MatchID            string `json:"match_id"`
	LeagueID           string `json:"league_id"`
	MatchDate          string `json:"match_date"`
	MatchTime          string `json:"match_time"`
	MatchStatus        string `json:"match_status"`
	MatchHometeamID    string `json:"match_hometeam_id"`
	MatchAwayteamID    string `json:"match_awayteam_id"`
	MatchHometeamScore string `json:"match_hometeam_score"`
	MatchAwayteamScore string `json:"match_awayteam_score"`
}

type apifootballError struct {
	Error   int   `json:"error"`
	Message string `json:"message"`
}

// FetchMatches calls get_events for the date range and league, then converts to entity.Match.
// Only matches where both home and away teams are in the mapping are included; others are skipped with a log.
// All requests use timezone=utc so match_date and match_time are in UTC.
func (p *Provider) FetchMatches(ctx context.Context, competition entity.Competition, from, to time.Time) ([]entity.Match, error) {
	leagueID, ok := CompetitionToAPIFootballID[competition]
	if !ok {
		return nil, fmt.Errorf("unsupported competition for apifootball: %d", competition)
	}

	fromStr := from.Format(time.DateOnly)
	toStr := to.Format(time.DateOnly)
	params := url.Values{
		"from":      {fromStr},
		"to":        {toStr},
		"league_id": {leagueID},
	}
	events, err := p.get(ctx, params)
	if err != nil {
		return nil, err
	}

	entityMatches := make([]entity.Match, 0, len(events))
	for _, ev := range events {
		match, err := eventToEntityMatch(ev, competition)
		if err != nil {
			slog.Debug("Skipping apifootball event (unmapped team or parse error)", "match_id", ev.MatchID, "error", err)
			continue
		}

		entityMatches = append(entityMatches, *match)
	}

	return entityMatches, nil
}

// FetchMatchByID fetches a single event by match_id (get_events with match_id param).
func (p *Provider) FetchMatchByID(ctx context.Context, providerMatchID string) (*entity.Match, error) {
	params := url.Values{"match_id": {providerMatchID}}
	events, err := p.get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match %s: %w", providerMatchID, err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no event found for match_id %s", providerMatchID)
	}

	competition, ok := APIFootballIDToCompetition[events[0].LeagueID]
	if !ok {
		return nil, fmt.Errorf("unknown league_id %s for match %s", events[0].LeagueID, providerMatchID)
	}

	match, err := eventToEntityMatch(events[0], competition)
	if err != nil {
		return nil, fmt.Errorf("failed to convert match %s: %w", providerMatchID, err)
	}

	return match, nil
}

// get calls the get_events endpoint, adding action, APIkey, and timezone=utc to the given params.
func (p *Provider) get(ctx context.Context, params url.Values) ([]apifootballEvent, error) {
	queryParams := make(url.Values)
	queryParams.Set("action", "get_events")
	queryParams.Set("APIkey", p.apiKey)
	queryParams.Set("timezone", "utc")
	for paramName, paramValue := range params {
		queryParams[paramName] = paramValue
	}
	rawURL := p.baseURL + "?" + queryParams.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("apifootball API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read get_events response: %w", err)
	}

	apiErr := &apifootballError{}
	if err := json.Unmarshal(body, apiErr); err == nil && apiErr.Error != 0 {
		if apiErr.Error == http.StatusNotFound {
			return []apifootballEvent{}, nil
		}

		if apiErr.Message != "" {
			return nil, fmt.Errorf("apifootball API error %d: %s", apiErr.Error, apiErr.Message)
		}

		return nil, fmt.Errorf("apifootball API error %d", apiErr.Error)
	}

	var events []apifootballEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("decode get_events response: %w", err)
	}

	return events, nil
}

func eventToEntityMatch(match apifootballEvent, competition entity.Competition) (*entity.Match, error) {
	if match.MatchHometeamID == "" || match.MatchAwayteamID == "" || match.MatchID == "" {
		return nil, fmt.Errorf("missing home or away team ID or match ID")
	}
	
	homeTeam, ok := APIFootballTeamMapping[match.MatchHometeamID]
	if !ok {
		return nil, fmt.Errorf("unmapped home team ID %s", match.MatchHometeamID)
	}

	awayTeam, ok := APIFootballTeamMapping[match.MatchAwayteamID]
	if !ok {
		return nil, fmt.Errorf("unmapped away team ID %s", match.MatchAwayteamID)
	}

	start, err := parseMatchStartUTC(match.MatchDate, match.MatchTime)
	if err != nil {
		return nil, fmt.Errorf("parse match_date/time: %w", err)
	}

	status := mapMatchStatus(match.MatchStatus)
	if status == entity.Finished && (match.MatchHometeamScore == "" || match.MatchAwayteamScore == "") {
		return nil, fmt.Errorf("missing home or away team score for finished match")
	}

	homeScore, err := parseScore(match.MatchHometeamScore)
	if err != nil {
		return nil, fmt.Errorf("parse home score %q: %w", match.MatchHometeamScore, err)
	}

	awayScore, err := parseScore(match.MatchAwayteamScore)
	if err != nil {
		return nil, fmt.Errorf("parse away score %q: %w", match.MatchAwayteamScore, err)
	}

	return entity.NewMatch(
		*start,
		entity.APIFootball,
		match.MatchID,
		homeTeam,
		awayTeam,
		homeScore,
		awayScore,
		competition,
		status,
	)
}

func parseMatchStartUTC(matchDate, matchTime string) (*time.Time, error) {
	if matchDate == "" || matchTime == "" {
		return nil, fmt.Errorf("missing date or time")
	}

	// adding seconds so we can work with time.DateTime instead of the weird formatting go has for Y-m-d H:i
	matchTime += ":00"

	dateTimeString := matchDate + " " + matchTime
	dateTime, err := time.ParseInLocation(time.DateTime, dateTimeString, time.UTC)
	if err != nil {
		return nil, err
	}

	return &dateTime, nil
}

func mapMatchStatus(matchStatus string) entity.MatchStatus {
	if matchStatus == "Finished" || matchStatus == "Awarded" || matchStatus == "After ET" || matchStatus == "After Pen." {
		return entity.Finished
	}

	if matchStatus == "Postponed" || matchStatus == "Cancelled" || matchStatus == "" {
		return entity.Pending
	}

	// This includes Half Time and the minute the match is currently in
	return entity.InProgress
}

func parseScore(score string) (uint, error) {
	if score == "" {
		return 0, nil
	}

	number, err := strconv.ParseUint(score, 10, 32)
	if err != nil {
		// TODO: we should log this error
		return 0, err
	}

	return uint(number), nil
}
