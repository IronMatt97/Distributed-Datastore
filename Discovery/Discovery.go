package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var MasterIP string = ""
var DSlist string = ""
var restAPIlist = list.New()

//TODO	IMPLEMENTA APICRASH
//TODO implementa che quando discovery rinasce riacquisisce le liste

func dsCrash(w http.ResponseWriter, r *http.Request) {
	dsToRemove := analyzeRequest(r)
	DSlist = strings.Replace(DSlist, dsToRemove+"|", "", -1) //Sostituisci sempre la stringa da rimuovere con spazio vuoto
	os.Remove("DS-" + dsToRemove)
}
func dsMasterCrash(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ds master crash was called")
	DSlist = strings.Replace(DSlist, MasterIP+"|", "", -1) //Sostituisci sempre la stringa da rimuovere con spazio vuoto
	os.Remove("DS-" + MasterIP)
	electNewMaster()
	requestJSON, _ := json.Marshal(MasterIP)
	http.Post("http://"+MasterIP+":8000/becomeMaster", "application/json", nil) //Avvisa il nuovo master che ora è master
	fmt.Println("I just told the new master he is new master now")
	for api := restAPIlist.Front(); api != nil; api = api.Next() {
		http.Post("http://"+fmt.Sprint(api)+":8000/changeMaster", "application/json", bytes.NewBuffer(requestJSON))
		fmt.Println("The new master is " + MasterIP + " and I am telling it to api :" + fmt.Sprint(api))
	}
}
func electNewMaster() {
	fmt.Println("The master has changed, old master was :" + MasterIP)
	for pos, char := range DSlist {
		fmt.Println(char)
		if char == 124 { //Quindi se ho trovato un |
			MasterIP = DSlist[0:pos]
			break
		}
	}
	fmt.Println("The master has changed, new master is :" + MasterIP)
}

func registerNewNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fmt.Println("Somebody registered:")
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
			MasterIP = dsIP
		}
		fmt.Println("I registered a new datastore: " + dsIP)
		requestJSON, _ := json.Marshal(dsIP)
		fmt.Println("STO AVVISANDO IL MASTER CHE CE UNA NUOVA REPLICA")
		http.Post("http://"+MasterIP+":8000/addDs", "application/json", bytes.NewBuffer(requestJSON)) //avvisa che c'è un nuovo DS
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
		response = MasterIP + "restAPI"
		fmt.Println("I registered a new restAPI: " + restAPI_IP)

	}
	//Answer requestOK
	json.NewEncoder(w).Encode(response)
	fmt.Println("I answered: " + response)
	fmt.Println("Lista delle API connesse: ")
	for api := restAPIlist.Front(); api != nil; api = api.Next() {
		fmt.Println(fmt.Sprint(api))
	}
	fmt.Println("Lista dei DS connessi: " + DSlist)

}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/register", registerNewNode).Methods("POST")
	router.HandleFunc("/dsCrash", dsCrash).Methods("POST")
	router.HandleFunc("/dsMasterCrash", dsMasterCrash).Methods("POST")
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
		if !strings.Contains(DSlist, ip) {
			DSlist = DSlist + ip + "|"
		}
	} else if mode == "restAPI" {
		var alreadyExists bool = false
		for api := restAPIlist.Front(); api != nil; api = api.Next() {
			if strings.Compare(fmt.Sprint(api.Value), ip) == 0 {
				alreadyExists = true
			}
		}
		if !alreadyExists {
			restAPIlist.PushBack(ip)
		}
	}
	return ip
}
