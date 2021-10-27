package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func registerNewNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	receivedRequest := analyzeRequest(r)
	if strings.Compare(receivedRequest, "datastore") == 0 {
		//Register new datastore
		dsIP := (r.Header.Get("X-REAL-IP"))
		fmt.Println(dsIP)
	}
	if strings.Compare(receivedRequest, "restAPI") == 0 {
		//Register new restAPI
	}
	//Answer requestOK
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/register", registerNewNode).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))
}
func analyzeRequest(r *http.Request) string {
	requestBody, err := ioutil.ReadAll(r.Body) //Read the request
	if err != nil {
		fmt.Println("An error has occurred trying to read client's request. ")
		fmt.Println(err.Error())
		return ""
	}
	var receivedRequest string                                  //Put client's request in a string
	err = json.Unmarshal([]byte(requestBody), &receivedRequest) //Unmarshal client's request
	if err != nil {
		fmt.Println("Error unmarshaling client's request.")
		fmt.Println(err.Error())
		return ""
	}
	return receivedRequest //Return unmarshaled string
}
