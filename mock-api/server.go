package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"mock_api/repository"
)

func Start() error {
	port := getPort()
	slog.Info("Starting HTTP server", "port", port)

	// Set up routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/competitions/", competitionsHandler)
	http.HandleFunc("/matches/", matchesHandler)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func getPort() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return port
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// GET /competitions/{id}/matches?dateFrom=...&dateTo=...
func competitionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /competitions/{id}/matches
	path := strings.TrimPrefix(r.URL.Path, "/competitions/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "matches" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	competitionID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid competition ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	dateFromStr := r.URL.Query().Get("dateFrom")
	dateToStr := r.URL.Query().Get("dateTo")

	if dateFromStr == "" || dateToStr == "" {
		http.Error(w, "Missing dateFrom or dateTo parameter", http.StatusBadRequest)
		return
	}

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		http.Error(w, "Invalid dateFrom format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		http.Error(w, "Invalid dateTo format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Query matches from database
	matches, err := repository.FindMatchesByDateRange(competitionID, dateFrom, dateTo)
	if err != nil {
		slog.Error("Failed to find matches", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := formatMatchesResponse(matches, competitionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /matches/{id}
func matchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /matches/{id}
	path := strings.TrimPrefix(r.URL.Path, "/matches/")
	matchID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	// Query match from database
	match, err := repository.FindMatchByID(matchID)
	if err != nil {
		slog.Error("Failed to find match", "error", err, "match_id", matchID)
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}

	response := formatMatch(*match)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func formatMatchesResponse(matches []repository.Match, competitionID int) map[string]interface{} {
	matchList := make([]map[string]interface{}, 0, len(matches))
	for _, m := range matches {
		matchList = append(matchList, formatMatch(m))
	}

	return map[string]interface{}{
		"filters": map[string]interface{}{},
		"resultSet": map[string]interface{}{
			"count":  len(matches),
			"first":  getFirstDate(matches),
			"last":   getLastDate(matches),
			"played": len(matches), // All matches are FINISHED
		},
		"competition": map[string]interface{}{
			"id":     competitionID,
			"name":   "Primera Division",
			"code":   "PD",
			"type":   "LEAGUE",
			"emblem": "https://crests.football-data.org/laliga.png",
		},
		"matches": matchList,
	}
}

func formatMatch(match repository.Match) map[string]interface{} {
	winner := ""
	if match.Status == "FINISHED" {
		if match.HomeTeamScore > match.AwayTeamScore {
			winner = "HOME_TEAM"
		} else if match.AwayTeamScore > match.HomeTeamScore {
			winner = "AWAY_TEAM"
		} else {
			winner = "DRAW"
		}
	}

	return map[string]interface{}{
		"id":          match.ID,
		"utcDate":     match.UTCDate,
		"status":      match.Status,
		"matchday":    match.Matchday,
		"lastUpdated": time.Now().Format("2006-01-02T15:04:05Z"),
		"homeTeam": map[string]interface{}{
			"id":        match.HomeTeamID,
			"name":      "Team " + fmt.Sprintf("%d", match.HomeTeamID),
			"shortName": "Team " + fmt.Sprintf("%d", match.HomeTeamID),
		},
		"awayTeam": map[string]interface{}{
			"id":        match.AwayTeamID,
			"name":      "Team " + fmt.Sprintf("%d", match.AwayTeamID),
			"shortName": "Team " + fmt.Sprintf("%d", match.AwayTeamID),
		},
		"score": map[string]interface{}{
			"winner":   winner,
			"duration": "REGULAR",
			"fullTime": map[string]interface{}{
				"home": match.HomeTeamScore,
				"away": match.AwayTeamScore,
			},
			"halfTime": map[string]interface{}{
				"home": nil,
				"away": nil,
			},
		},
	}
}

func getFirstDate(matches []repository.Match) string {
	if len(matches) == 0 {
		return ""
	}

	if t, err := time.Parse("2006-01-02T15:04:05Z", matches[0].UTCDate); err == nil {
		return t.Format("2006-01-02")
	}

	return ""
}

func getLastDate(matches []repository.Match) string {
	if len(matches) == 0 {
		return ""
	}

	// Parse the last match's date
	last := matches[len(matches)-1]
	if t, err := time.Parse("2006-01-02T15:04:05Z", last.UTCDate); err == nil {
		return t.Format("2006-01-02")
	}

	return ""
}
