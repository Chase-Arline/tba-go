package main

import (
	"fmt"

	"github.com/Chase-Arline/tba-go/tba"
)

func main() {
	client := new(tba.TBAClient)
	event, err := client.FetchEvent("Yakima", "Sundome", "Yakima Sundome", 2018)
	errHandler(err)
	_, err = client.FetchEventStatistics(event)
	errHandler(err)

	qms := client.GetSortedQMs(event)
	//allOPRs := make(map[string][]float64)
	fmt.Println(qms[0])
	fmt.Println(qms[1])
	fmt.Println(qms[2])
	oprs, _ := client.GenerateOPRs(qms[0:2])
	fmt.Println(oprs["frc3218"])
	//p := plot.New() //line that enables gonum plot import

}

func errHandler(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}
