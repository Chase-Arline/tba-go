package tba

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
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
