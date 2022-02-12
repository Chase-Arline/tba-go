package tba

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"gonum.org/v1/gonum/mat"
)

//Client represents an abstracted HTTP Client to pull from TBA API
type TBAClient struct {
	http.Client
}

var year int = time.Now().Year()

//Request is the struct used for making a request on the blue alliance API
type Request struct {
	http.Request
	url string
}

const baseURL string = "https://www.thebluealliance.com/api/v3"
const httpKey string = "o1JhCWWwjfMbtKos9WBVuK6HR6H98KWrT7VUuQgWFdAF5kvnLSvWxmuYxFbswk1H "

func errHandler(err error) {
	if err != nil {
		panic(err)
	}
}

//Request is the default method for sending data requests to TBA
func (c TBAClient) Request(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", baseURL+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-TBA-Auth-Key", httpKey)
	return c.Do(req)
}

//NewClient returns a client used for pulling data from TBA at the year specified
func NewClient(year int, city string) TBAClient {
	c := TBAClient{}
	c.Timeout = 10 * time.Second
	return c
}

//Team gives the string key representation of an frc team
func Team(team int) (teamKey string) {
	return fmt.Sprintf("%s%v", "frc", team)
}

//Event represents the information wanted from the http response - underscores are used for JSON parsing
type Event struct {
	Key, Name, Event_code, City, State_prov, Short_name, Location_name string
	Year                                                               int
}

//ErrEventNotFound is returned when keywords are not able to match to an event or an event code is not available on TBA
var ErrEventNotFound error = errors.New("event could not be found")

//FetchEvent returns the string key representation of an frc event
//examples of keywords are city, event name, location name, short name of the event, ~ Auburn, Auburn Mountainview, PNW
func (c TBAClient) FetchEvent(location, city, name string, year int) (Event, error) {
	response, err := c.Request(fmt.Sprintf("%s%v", "/events/", year))
	if err != nil {
		return Event{}, err
	}
	events := []Event{}
	if err = bodyToJSON(response, &events); err != nil {
		return Event{}, err
	}
	var match Event
	var highestMatch int
	for _, event := range events {
		thisMatch := 0
		if strings.Contains(event.City, city) {
			thisMatch++
		}
		if strings.Contains(event.Location_name, location) {
			thisMatch++
		}
		if strings.Contains(event.Name, name) {
			thisMatch++
		}
		if strings.Contains(event.Short_name, name) {
			thisMatch++
		}
		if thisMatch > highestMatch {
			match = event
			highestMatch = thisMatch
		}
	}
	if highestMatch == 0 {
		return Event{}, ErrEventNotFound
	}
	return match, nil
}

func (c TBAClient) FetchEventStatistics(e Event) (es EventStatistics, err error) {
	response, err := c.Request(fmt.Sprintf("%s", "/event/"+e.Key+"/oprs"))
	if err != nil {
		return EventStatistics{}, err
	}
	es = EventStatistics{}
	if err = bodyToJSON(response, &es); err != nil {
		return es, err
	}
	return
}

func (c TBAClient) FetchMatches(e Event) (matches []Match, err error) {
	r, err := c.Request("/event/" + e.Key + "/matches/simple")
	if err != nil {
		return
	}
	matches = []Match{}
	if err = bodyToJSON(r, &matches); err != nil {
		return
	}
	return
}

func (c TBAClient) FetchTeamStatuses(e Event, team int) (ts []TeamEventStatus, err error) {
	r, err := c.Request("/team/" + teamKey(team) + "/event/" + e.Key + "/status")
	if err != nil {
		return
	}
	ts = []TeamEventStatus{}
	if err = bodyToJSON(r, &ts); err != nil {
		return
	}
	fmt.Println(ts)
	return
}

func teamKey(n int) string {
	return fmt.Sprintf("%s%v", "frc", n)
}

func createMatrix(matches []Match, teamToCol map[string]int, colToTeam map[int]string) (matrix *mat.Dense, vector *mat.VecDense) {
	matrix = mat.NewDense(2*len(matches), len(teamToCol), nil)
	vector = mat.NewVecDense(2*len(matches), nil)
	for i, match := range matches {
		matrix.Set(i*2, teamToCol[match.Alliances.Blue.Team_keys[0]], 1)
		matrix.Set(i*2, teamToCol[match.Alliances.Blue.Team_keys[1]], 1)
		matrix.Set(i*2, teamToCol[match.Alliances.Blue.Team_keys[2]], 1)
		matrix.Set(i*2+1, teamToCol[match.Alliances.Red.Team_keys[0]], 1)
		matrix.Set(i*2+1, teamToCol[match.Alliances.Red.Team_keys[1]], 1)
		matrix.Set(i*2+1, teamToCol[match.Alliances.Red.Team_keys[2]], 1)
		vector.SetVec(i*2, float64(match.Alliances.Blue.Score))
		vector.SetVec(i*2+1, float64(match.Alliances.Red.Score))
	}
	return
}

func solveMatrix(matrix *mat.Dense, vector *mat.VecDense) (vec *mat.VecDense, err error) {
	_, c := matrix.Dims()
	vec = mat.NewVecDense(c, nil)
	err = vec.SolveVec(matrix, vector)
	errHandler(err)
	return
}

func makeBiMap(matches []Match) (teamToCol map[string]int, colToTeam map[int]string, err error) {
	var teamNum int = 0
	teamSet := make(map[string]struct{})
	teamToCol = make(map[string]int)
	colToTeam = make(map[int]string)
	var exists = struct{}{}
	for _, match := range matches {
		for _, blueTeam := range match.Alliances.Blue.Team_keys {
			if _, ok := teamSet[blueTeam]; !ok {
				teamSet[blueTeam] = exists
				teamToCol[blueTeam] = teamNum
				colToTeam[teamNum] = blueTeam
				teamNum++
			}
		}
		for _, redTeam := range match.Alliances.Red.Team_keys {
			if _, ok := teamSet[redTeam]; !ok {
				teamSet[redTeam] = exists
				teamToCol[redTeam] = teamNum
				colToTeam[teamNum] = redTeam
				teamNum++
			}
		}
	}
	return
}

func sortMatches(matches []Match) error {
	matchLess := func(i, j int) bool {
		return matches[i].Match_number < matches[j].Match_number
	}
	sort.SliceStable(matches, matchLess)
	return nil
}

func (client *TBAClient) GetSortedQMs(event Event) []Match {
	matches, err := client.FetchMatches(event)
	errHandler(err)
	qms := make([]Match, 0, len(matches))
	for _, match := range matches {
		if match.Comp_level == "qm" {
			qms = append(qms, match)
		}
	}
	err = sortMatches(qms)
	errHandler(err)
	return matches
}

func (client *TBAClient) GenerateOPRs(qms []Match) (map[string]float64, []string) {
	oprs := make(map[string]float64)
	teamToCol, colToTeam, err := makeBiMap(qms)
	errHandler(err)
	matrix, vector := createMatrix(qms, teamToCol, colToTeam)
	vec, err := solveMatrix(matrix, vector)
	errHandler(err)
	teams := make([]string, len(teamToCol))
	var i int = 0
	for col, team := range colToTeam {
		oprs[team] = vec.AtVec(col)
		teams[i] = team
		i++
	}
	return oprs, teams
}
