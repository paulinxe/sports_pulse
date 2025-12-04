package football_org

type MatchesResponse struct {
	Matches []Match `json:"matches"`
}

type Match struct {
	ID       int    `json:"id"`
	UTCDate  string `json:"utcDate"`
	Status   string `json:"status"`
	HomeTeam Team   `json:"homeTeam"`
	AwayTeam Team   `json:"awayTeam"`
	Score    Score  `json:"score"`
}

type Team struct {
	ID int `json:"id"`
}

type Score struct {
	FullTime ScoreTime `json:"fullTime"`
}

type ScoreTime struct {
	Home int `json:"home"`
	Away int `json:"away"`
}
