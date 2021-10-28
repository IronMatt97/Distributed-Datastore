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

var MasterIP string = ""
var DSlist string = ""
var restAPIlist string = ""

func registerNewNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	receivedRequest := analyzeRequest(r)
	response := "ok"
	if strings.Compare(receivedRequest, "datastore") == 0 {
		//Register new datastore
		dsIP := acquireIP(r.RemoteAddr, "datastore")            //Aggiungi alla lista di ip e restituiscilo
		err := ioutil.WriteFile("DS-"+dsIP, []byte(dsIP), 0777) //Write the file
		if err != nil {
			fmt.Println("An error has occurred trying to register the datastore. ")
			fmt.Println(err.Error())
			return
		}
		if MasterIP == "" {
			response = DSlist + "master"
		}
	}
	if strings.Compare(receivedRequest, "restAPI") == 0 {
		//Register new restAPI
		restAPI_IP := acquireIP(r.RemoteAddr, "restAPI")                     //Aggiungi alla lista di ip e restituiscilo
		err := ioutil.WriteFile("API-"+restAPI_IP, []byte(restAPI_IP), 0777) //Write the file
		if err != nil {
			fmt.Println("An error has occurred trying to register the datastore. ")
			fmt.Println(err.Error())
			return
		}
	}
	//Answer requestOK
	json.NewEncoder(w).Encode(response)
	//TODO send to master new list every time
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
func acquireIP(ip string, mode string) string {

	ip = ip[0:len(ip)-6] + "" //RITAGLIA L'IP

	if mode == "datastore" {
		if strings.Contains(DSlist, ip) == false {
			DSlist = DSlist + ip + "|"
		}
	} else if mode == "restAPI" {
		if strings.Contains(restAPIlist, ip) == false {
			restAPIlist = restAPIlist + ip + "|"
		}
	}
	return ip
}
