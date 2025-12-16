package main

import (
	"math/rand"
	"time"
)

type ScheduledMatch struct {
	HomeTeamID uint
	AwayTeamID uint
	Matchday   int
	Date       time.Time
}

// GenerateSchedule generates matches for the next 5 weekends starting from now
// Returns 50 matches (5 weekends × 10 matches per weekend)
func GenerateSchedule() ([]ScheduledMatch, error) {
	// These are the football_org API team IDs
	teamIDs := []uint{
		263,  // Alaves
		77,   // AthleticClub
		78,   // AtleticoMadrid
		81,   // Barcelona
		558,  // CeltaVigo
		285,  // Elche
		80,   // Espanyol
		82,   // Getafe
		298,  // Girona
		88,   // Levante
		89,   // Mallorca
		79,   // Osasuna
		1048, // Oviedo
		87,   // RayoVallecano
		90,   // RealBetis
		86,   // RealMadrid
		92,   // RealSociedad
		559,  // Sevilla
		95,   // Valencia
		94,   // Villarreal
	}

	matches := generateMatches(teamIDs, 5)

	now := time.Now()
	assignMatchDates(matches, now)

	return matches, nil
}

func generateMatches(teams []uint, weekends int) []ScheduledMatch {
	const matchesPerWeekend = 10
	totalMatches := weekends * matchesPerWeekend

	var matches []ScheduledMatch

	// Generate all possible pairs
	var pairs []struct{ home, away uint }
	for i := 0; i < len(teams); i++ {
		for j := 0; j < len(teams); j++ {
			if i != j {
				pairs = append(pairs, struct{ home, away uint }{teams[i], teams[j]})
			}
		}
	}

	// Shuffle pairs to randomize
	rand.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})

	// Take the first 'totalMatches' pairs
	for i := 0; i < totalMatches && i < len(pairs); i++ {
		matches = append(matches, ScheduledMatch{
			HomeTeamID: pairs[i].home,
			AwayTeamID: pairs[i].away,
			Matchday:   (i / matchesPerWeekend) + 1, // Matchday 1-weekends
		})
	}

	return matches
}


// Distributes matches weekly (weekends), with random times (16:00, 18:00, 20:00 UTC)
func assignMatchDates(matches []ScheduledMatch, now time.Time) {
	// Group matches by matchday (1-5)
	matchdayGroups := make(map[int][]*ScheduledMatch)
	for i := range matches {
		matchdayGroups[matches[i].Matchday] = append(matchdayGroups[matches[i].Matchday], &matches[i])
	}

	// Time slots: 16:00, 18:00, 20:00 UTC
	timeSlots := []int{16, 18, 20}

	currentDate := findNextSaturday(now)

	// Process 5 matchdays (5 weekends)
	for matchday := 1; matchday <= 5; matchday++ {
		matchesInDay, exists := matchdayGroups[matchday]
		if !exists {
			continue
		}

		// Distribute matches across Saturday and Sunday (10 matches per matchday)
		// Split: 5 on Saturday, 5 on Sunday
		saturdayMatches := matchesInDay[:5]
		sundayMatches := matchesInDay[5:10]

		// Assign Saturday matches
		for _, match := range saturdayMatches {
			hour := timeSlots[rand.Intn(len(timeSlots))]
			matchDate := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), hour, 0, 0, 0, time.UTC)
			match.Date = matchDate
		}

		// Assign Sunday matches
		sundayDate := currentDate.Add(24 * time.Hour)
		for _, match := range sundayMatches {
			hour := timeSlots[rand.Intn(len(timeSlots))]
			matchDate := time.Date(sundayDate.Year(), sundayDate.Month(), sundayDate.Day(), hour, 0, 0, 0, time.UTC)
			match.Date = matchDate
		}

		// Move to next weekend (7 days later)
		currentDate = currentDate.Add(7 * 24 * time.Hour)
	}
}

func findNextSaturday(from time.Time) time.Time {
	daysUntilSaturday := (int(time.Saturday) - int(from.Weekday()) + 7) % 7

	if daysUntilSaturday == 0 {
		daysUntilSaturday = 7 // If it's already Saturday, go to next Saturday
	}

	return from.Add(time.Duration(daysUntilSaturday) * 24 * time.Hour)
}
