package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	params := mux.Vars(r)                       //Acquire url params
	data, err := ioutil.ReadFile(params["key"]) //Try to read the requested file
	if err != nil {
		fmt.Println("An error has occurred reading the file.")
		fmt.Println(err.Error())
		return
	}
	json.NewEncoder(w).Encode(string(data)) //Send the response to the client
}

func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	receivedRequest := analyzeRequest(r)
	var info []string = strings.Split(receivedRequest, "|") //Acquire file name and content from client's request
	var fileName string = info[0]
	var fileContent string = info[1]

	if _, err := os.Stat(fileName); err == nil {
		json.NewEncoder(w).Encode("The file you requested already exists.") //Return error if file already exists
		return
	}

	err := ioutil.WriteFile(fileName, []byte(fileContent), 0777) //Write the file
	if err != nil {
		fmt.Println("An error has occurred trying to write the file. ")
		fmt.Println(err.Error())
		return
	}
	json.NewEncoder(w).Encode("The file was successfully uploaded.")
}

func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fileToRemove := analyzeRequest(r)
	err := os.Remove(fileToRemove) // Remove the file
	if err != nil {
		fmt.Println("An error has occurred trying to delete the file.")
		fmt.Println(err.Error())
		return
	}
	json.NewEncoder(w).Encode("The file was successfully removed.")
}

func main() {
	router := mux.NewRouter()                           //Router initialization
	router.HandleFunc("/put", put).Methods("POST")      //put requests handler/endpoint
	router.HandleFunc("/get/{key}", get).Methods("GET") //get requests handler/endpoint
	router.HandleFunc("/delete", del).Methods("POST")   //del requests handler/endpoint
	log.Fatal(http.ListenAndServe(":8000", router))     //Listen and serve requests on port 8000
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
