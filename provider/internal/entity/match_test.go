package entity

import (
	"testing"
	"time"
)

// As we are emulating the abi.encodePacked Solidity function and we are using it with more than 1 parameter,
// we are just making sure that we don't have collisions when generating the match ID.
// As we know, it can lead to collisions when using strings as parameters which is not the case here but hey, tests are free.
// string memory s1 = "1";
// string memory s2 = "10";
//
// string memory s3 = "11";
// string memory s4 = "0";

// abi.encodePacked(s1, s2); // "1" + "10" = "110"
// abi.encodePacked(s3, s4); // "11" + "0" = "110"  // COLLISION!
func Test_we_dont_have_collisions_when_generating_match_id(t *testing.T) {
	score := uint(0)
	match1, err := NewMatch(
		time.Now(),
		FootballOrg,
		"1234567890",
		1,
		11,
		score,
		score,
		LaLiga,
		Pending,
	)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	match2, err := NewMatch(
		time.Now(),
		FootballOrg,
		"1234567890",
		11,
		1,
		score,
		score,
		LaLiga,
		Pending,
	)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	match3, err := NewMatch(
		time.Now(),
		FootballOrg,
		"1234567890",
		1,
		11,
		score,
		score,
		111,
		Pending,
	)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	match4, err := NewMatch(
		time.Now(),
		FootballOrg,
		"1234567890",
		11,
		1,
		score,
		score,
		111,
		Pending,
	)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	ids := []string{
		match1.CanonicalID,
		match2.CanonicalID,
		match3.CanonicalID,
		match4.CanonicalID,
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	if len(unique) != len(ids) {
		t.Errorf("Expected %d unique IDs, but got %d. We have collisions", len(ids), len(unique))
	}
}
