package db

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type leader struct {
	Name string
	Wins int
}

// GetLeaderBoard returns a json array of the current top 10 winners
func GetLeaderBoard(w http.ResponseWriter, r *http.Request) {
	rows, err := paoDb.Query("select winner, count(*) from completedgames where winner != '' group by winner order by count(*) desc limit 10")
	if err != nil {
		fmt.Printf("Could not retrieve leader board: %s\n", err)
	}
	leaders := []leader{}
	for rows.Next() {
		var l leader
		err = rows.Scan(&l.Name, &l.Wins)
		if err != nil {
			fmt.Printf("Error getting leader board: %s\n", err)
		}
		leaders = append(leaders, l)
		fmt.Printf("Got a leader: %+v\n", l)
	}
	js, err := json.Marshal(leaders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(string(js))
	w.Write(js)
}
