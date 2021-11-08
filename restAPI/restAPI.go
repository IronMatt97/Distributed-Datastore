package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var DiscoveryIP string = "172.17.0.2"
var DSMasterIP string = ""
var DSBalancerIP string = "" //Ricordati nella funzione get di rimetterci DSBALANCERIp quando lo passi su amazon

func reportDSMasterCrash() {
	fmt.Println("Master crashed: sending this to discovery.")
	var request string = DSMasterIP //Build the request in a particular format
	requestJSON, _ := json.Marshal(request)
	_, err := http.Post("http://"+DiscoveryIP+":8080/dsMasterCrash", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	for err != nil {
		fmt.Println("discovery crashed. Waitng for it to restart.")
		time.Sleep(3 * time.Second)
		_, err = http.Post("http://"+DiscoveryIP+":8080/dsMasterCrash", "application/json", bytes.NewBuffer(requestJSON))
	}
}
func changeDSMasterOnCrash(w http.ResponseWriter, r *http.Request) {
	DSMasterIP = analyzeRequest(r)
	fmt.Println("Master crashed: the new master is " + DSMasterIP)
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	params := mux.Vars(r)                                                            //RIMETTI QUI DSBalancerIP TODO
	fmt.Println("get called: I wanna read " + params["key"] + " on " + DSMasterIP)   //Acquire url params
	response, err := http.Get("http://" + DSMasterIP + ":8080/get/" + params["key"]) //Submitting a get request
	if err != nil {
		reportDSMasterCrash()
		fmt.Println(err.Error()) //Questa cosa qui andrà cambiata, se viene usato il load balancer in teoria non si verifica mai
		//perche ci sarà sempre una copia su che potrà fornire.
		fmt.Println("An error has occurred trying to estabilish a connection with the Datastore. Retry later.")
		fmt.Println(err.Error())
		return
	}
	responseFromDS, err := ioutil.ReadAll(response.Body) //Receiving http response
	if err != nil {
		fmt.Println("An error has occurred trying to read the requested file.")
		fmt.Println(err.Error())
		return
	}
	fmt.Println("the file requested is " + string(responseFromDS))
	json.NewEncoder(w).Encode(string(responseFromDS)) //Send the response to the client
}

func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	receivedRequest := analyzeRequest(r)
	var info []string = strings.Split(receivedRequest, "|") //Acquire file name and content from client's request
	var fileName string = info[0]
	var fileContent string = info[1]
	var request string = fileName + "|" + fileContent //Build the request in a particular format
	fmt.Println("put called: I wanna write " + request + " on " + DSMasterIP)
	requestJSON, _ := json.Marshal(request)
	response, err := http.Post("http://"+DSMasterIP+":8080/put", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	if err != nil {
		reportDSMasterCrash()
		fmt.Println(err.Error())
		json.NewEncoder(w).Encode("Master crashed. Try again later.")
		return
	}
	responseFromDS, err := ioutil.ReadAll(response.Body) //Receiving http response
	json.NewEncoder(w).Encode(string(responseFromDS))
}

func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fileToRemove := analyzeRequest(r)
	var request string = fileToRemove //Build the request in a particular format
	fmt.Println("del called: I wanna remove " + fileToRemove + " on " + DSMasterIP)
	requestJSON, _ := json.Marshal(request)
	response, err := http.Post("http://"+DSMasterIP+":8080/del", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	if err != nil {
		reportDSMasterCrash()
		fmt.Println(err.Error())
		json.NewEncoder(w).Encode("Master crashed. Try again later.")
		return
	}
	responseFromDS, err := ioutil.ReadAll(response.Body) //Receiving http response
	json.NewEncoder(w).Encode(string(responseFromDS))
}

func main() {
	register()
	router := mux.NewRouter()                           //Router initialization
	router.HandleFunc("/put", put).Methods("POST")      //put requests handler/endpoint
	router.HandleFunc("/get/{key}", get).Methods("GET") //get requests handler/endpoint
	router.HandleFunc("/del", del).Methods("POST")      //del requests handler/endpoint
	router.HandleFunc("/changeMaster", changeDSMasterOnCrash).Methods("POST")
	router.HandleFunc("/whoIsMaster", whoIsMaster).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router)) //Listen and serve requests on port 8080
}
func whoIsMaster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fmt.Println("Il discovery mi ha chiesto chi è il master, e per me è " + DSMasterIP)
	json.NewEncoder(w).Encode(DSMasterIP)
}
func register() {
	requestJSON, _ := json.Marshal("restAPI")
	response, err := http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Se fallisce riprova ogni 3 secondi
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node.")
		fmt.Println(err.Error())
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body) //Receiving http response
	//l'API NON PUO REGISTRARSI FINCHE NON CE UN MASTER------- master è "" se non esiste! controlli e lo fai ripartire, deve aspettare finche non c'è un master!
	fmt.Println("The discovery answered: " + string(responseFromDiscovery))
	for strings.Compare(string(responseFromDiscovery)[1:len(string(responseFromDiscovery))-2], "restAPI") == 0 {
		fmt.Println("The master is not here yet. I am gonna wait")
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
		responseFromDiscovery, _ = ioutil.ReadAll(response.Body) //Receiving http response
	}

	DSMasterIP = (string(responseFromDiscovery[1 : len(string(responseFromDiscovery))-9]))
	DSMasterIP = strings.ReplaceAll(DSMasterIP, "\\", "")
	DSMasterIP = strings.ReplaceAll(DSMasterIP, "n", "") //Cleaning the output
	DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
	fmt.Println("registration complete: the master is" + DSMasterIP)
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
