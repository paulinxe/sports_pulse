package api

type MatchesResponse struct {
	Matches []FootballOrgMatch `json:"matches"`
}

type FootballOrgMatch struct {
	ID       uint   `json:"id"`
	UTCDate  string `json:"utcDate"`
	Status   string `json:"status"`
	HomeTeam Team   `json:"homeTeam"`
	AwayTeam Team   `json:"awayTeam"`
	Score    Score  `json:"score"`
}

type Team struct {
	ID uint `json:"id"`
}

type Score struct {
	FullTime ScoreTime `json:"fullTime"`
}

type ScoreTime struct {
	Home uint `json:"home"`
	Away uint `json:"away"`
}

// The happy path is to have a match in status FINISHED.
// If a match gets cancelled and never gets played, it will be in status AWARDED.
func (match *FootballOrgMatch) IsInFinalStatus() bool {
	return match.Status == "FINISHED" || match.Status == "AWARDED"
}

// Statuses: IN_PLAY, PAUSED, SUSPENDED indicate matches that are actively in progress.
// Note: TIMED, SCHEDULED are not in-progress (match hasn't started yet).
func (match *FootballOrgMatch) IsInProgress() bool {
	inProgressStatuses := map[string]bool{
		"IN_PLAY":   true,
		"PAUSED":    true,
		"SUSPENDED": true,
	}
	return inProgressStatuses[match.Status]
}
