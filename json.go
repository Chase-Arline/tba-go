package tba

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

//EventStatistics holds the OPR, DPR, and CCWM information for each team at an event
type EventStatistics struct {
	OPRS  map[string]float64
	DPRS  map[string]float64
	CCWMS map[string]float64
}

//TeamEventStatus holds qualification, alliance, and playoff data for a single team at an event
type TeamEventStatus struct {
	Qual     QualificationData
	Alliance AllianceData
	Playoff  PlayoffData
	Alliance_status_str, Playoff_status_str, Overall_status_str,
	next_match_key, last_match_key string
}

//QualificaticationData represents the TBA qualification data
type QualificationData struct {
	Num_teams int
	Ranking   []RankingData
	Status    string
}

//RankingData represents the TBA data in the qualification struct
type RankingData struct {
	Matches_played, rank, dq int
	Qual_average             float64
	Record                   BasicRecord
	Team_key                 string
}

//BasicRecord represents the Win-Loss-Tie record
type BasicRecord struct {
	Losses, Wins, Ties int
}

//AllianceData represents some small data about a team's alliance
type AllianceData struct {
	Name         string
	Number, Pick int
}

type PlayoffData struct {
	level, status        string //status is enum with won,eliminated playing iota
	playoff_average      int
	Current_level_record BasicRecord
	Record               BasicRecord
}

func bodyToJSON(h *http.Response, v interface{}) error {
	b, err := ioutil.ReadAll(h.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
