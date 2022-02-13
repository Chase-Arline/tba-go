package main

import (
	"fmt"
	"os"

	"github.com/Chase-Arline/tba-go/tba"
	"gonum.org/v1/plot"
)

func main() {
	client := new(tba.TBAClient)
	event, err := client.FetchEvent("Auburn", "Mountainview", "Auburn Mountainview", 2019)
	errHandler(err)
	_, err = client.FetchEventStatistics(event)
	errHandler(err)

	qms := client.GetSortedQMs(event)
	allOPRs := make(map[string][]float64)
	for i := 1; i < len(qms); i++ { //update oprs every qualifcation match that occurs
		oprs, teams := client.GenerateOPRs(qms[0:i])
		for _, team := range teams {
			allOPRs[team] = append(allOPRs[team], oprs[team]) //set each opr for each team up to this qualifcation match
		}
	}

	var userInput string
	for {
		fmt.Println("What team would you like to see graphed? Example: 3218\nType 'quit' to end the program. ")
		fmt.Scanf("%s", &userInput)
		if userInput == "quit" {
			break
		}
		realMatchNumberOffset := len(qms) - len(allOPRs["frc"+userInput])
		p := plot.New() //line that enables gonum plot import
		line, err := tba.DrawStatLine(p, allOPRs["frc"+userInput], realMatchNumberOffset)
		errHandler(err)
		p.Add(line)
		err = os.Mkdir("out", 0755)
		errHandler(err)
		err = p.Save(1920, 1080, "out/"+event.Name+" "+userInput+".png")
		errHandler(err)
		fmt.Printf("Saved as: %v\n", "out/ "+event.Name+" "+userInput+".png")
	}

}

func errHandler(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}
