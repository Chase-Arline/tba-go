package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Chase-Arline/tba-go/tba"
	"gonum.org/v1/plot"
)

func main() {
	client := new(tba.TBAClient)
	event, err := client.FetchEvent("Glacier", "Peak", "Glacier Peak", 2018)
	errHandler(err)
	_, err = client.FetchEventStatistics(event)
	errHandler(err)

	qms := client.GetSortedQMs(event)

	allOPRs := make(map[string][]float64)
	for i := 1; i < len(qms); i++ { //update oprs every qualifcation match that occurs
		oprs, teams, err := client.GenerateOPRs(qms[0:i])
		if err == nil {
			for _, team := range teams {
				allOPRs[team] = append(allOPRs[team], oprs[team]) //set each opr for each team up to this qualifcation match
			}
		} else {
			for _, team := range teams {
				allOPRs[team] = append(allOPRs[team], tba.SINGULAR_OPR) //bad opr is used for match placeholder and not used later when graphing
			}
		}
	}
	var userInput string
	for {
		fmt.Println("What team would you like to see graphed? Example: 3218\nType 'quit' to end the program. ")
		fmt.Scanf("%s", &userInput)
		if userInput == "quit" {
			break
		} else if _, ok := allOPRs["frc"+userInput]; !ok {
			fmt.Println("Input or team number not recognized. Was the team present in this competition: ", event.Name, "\nPlease try again")
		} else { //do graphing
			realMatchNumberOffset := len(qms) - len(allOPRs["frc"+userInput])
			p := plot.New() //line that enables gonum plot import
			line, err := tba.DrawStatLine(p, allOPRs["frc"+userInput], realMatchNumberOffset)
			errHandler(err)
			p.Add(line)
			err = os.Mkdir("out", 0755)
			if err != nil && !errors.Is(err, os.ErrExist) {
				errHandler(err)
			}
			err = p.Save(1920, 1080, "out/"+event.Name+" "+userInput+".png")
			errHandler(err)
			fmt.Printf("Saved as: %v\n", "out/ "+event.Name+" "+userInput+".png")
		}
	}

}

func errHandler(err error) {
	if err != nil {
		fmt.Println(err.Error)
		panic(err)
	}
}
