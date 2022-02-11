package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Chase-Arline/tba/tba"
	"gonum.org/v1/gonum/mat"
)

func main() {
	client := new(tba.TBAClient)
	event, err := client.FetchEvent("Auburn", "Mountainview", "Auburn Mountainview", 2019)
	errHandler(err)
	matches, err := client.FetchMatches(event)
	errHandler(err)
	fmt.Printf("Matches Length: %v\n ", len(matches))
	qms := make([]tba.Match, 0, len(matches))
	for _, match := range matches {
		if match.Comp_level == "qm" {
			qms = append(qms, match)
		}
	}
	fmt.Printf("QMs Length: %v\n", len(qms))
	fmt.Println("Sorting Matches")
	err = sortMatches(qms, true)
	errHandler(err)
	fmt.Println("Matches Sorted")
	teamToCol, colToTeam, err := makeBiMap(qms)
	fmt.Println(len(teamToCol))
	for _, match := range qms {
		fmt.Printf("Match #: %v\tRed: %v\t Blue: %v\n", match.Match_number, match.Alliances.Red.Score, match.Alliances.Blue.Score)
	}

	matrix, vector := createMatrix(qms, teamToCol, colToTeam)
	err = solveMatrix(matrix, vector)

}

func createMatrix(matches []tba.Match, teamToCol map[string]int, colToTeam map[int]string) (matrix *mat.Dense, vector *mat.VecDense) {
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

func solveMatrix(matrix *mat.Dense, vector *mat.VecDense) (err error) {
	r, c := matrix.Dims()
	resultVector := mat.NewVecDense(c, nil)
	fmt.Printf("matrix R: %v\t C: %v\n", r, c)
	fmt.Printf("ScoreVector: %v\n", vector.Len())
	fmt.Printf("Result Vector: %v\n", resultVector.Len())
	err = resultVector.SolveVec(matrix, vector)
	errHandler(err)
	fmt.Println(resultVector.RawVector().Data)
	return err
}

func makeBiMap(matches []tba.Match) (teamToCol map[string]int, colToTeam map[int]string, err error) {
	var teamNum int = 0
	teamSet := make(map[string]struct{})
	teamToCol = make(map[string]int)
	colToTeam = make(map[int]string)
	var exists = struct{}{}
	fmt.Println(len(matches))
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

func sortMatches(matches []tba.Match, test bool) error {
	matchLess := func(i, j int) bool {
		return matches[i].Match_number < matches[j].Match_number
	}
	sort.SliceStable(matches, matchLess)
	if test {
		var lastMatchN int = -1
		for _, match := range matches {
			if lastMatchN > match.Match_number {
				return errors.New("sorting issue: lower index has higher match number")
			}
			lastMatchN = match.Match_number
		}
	}
	return nil
}

func errHandler(err error) {
	if err != nil {
		fmt.Println(err.Error)
		panic(err)
	}
}
