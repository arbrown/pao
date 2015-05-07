package main

import "fmt"
import "net/http"

func main() {
	fmt.Printf("Hello, Pao\n")
	http.Handle("/", http.FileServer(http.Dir(".")))

	err := http.ListenAndServe(":2015", nil)
	if err != nil {
		panic("ListenAndServe:" + err.Error())
	}
}
