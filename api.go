package tba

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"tbajson"
	"time"
)

//Client represents an abstracted HTTP Client to pull from TBA API
type client struct {
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

func main() {
	client := new(client)
	event, err := client.FetchEvent("Auburn", "Auburn", "Mountainview", 2019)
	if err != nil {
		panic(err)
	}
	stats, err := client.FetchEventStatistics(event)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(stats.OPRS))
}

//Request is the default method for sending data requests to TBA
func (c client) Request(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", baseURL+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-TBA-Auth-Key", httpKey)
	return c.Do(req)
}

//NewClient returns a client used for pulling data from TBA at the year specified
func NewClient(year int, city string) client {
	c := client{}
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
func (c client) FetchEvent(location, city, name string, year int) (Event, error) {
	response, err := c.Request(fmt.Sprintf("%s%v", "/events/", year))
	if err != nil {
		return Event{}, err
	}
	events := []Event{}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return Event{}, err
	}
	err = json.Unmarshal(b, &events)
	if err != nil {
		fmt.Println(err)
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

func (c client) FetchEventStatistics(e Event) (es tbajson.EventStatistics, err error) {
	response, err := c.Request(fmt.Sprintf("%s", "/event/"+e.Key+"/oprs"))
	if err != nil {
		return EventStatistics{}, err
	}
	es = EventStatistics{}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(b, &es); err != nil {
		return es, err
	}
	return
}

func (c client) FetchTeamStatuses(e Event) (ts []TeamEventStatus, err error) {

}

func teamKey(n int) string {
	return fmt.Sprintf("%s%v", "frc", n)
}
