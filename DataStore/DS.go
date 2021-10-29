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
	"time"

	"github.com/gorilla/mux"
)

var DiscoveryIP = "localhost"
var Master bool = false
var DSList = list.New()

func put(w http.ResponseWriter, r *http.Request) {
	//Aggiorno me stesso
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

	//Se sono master aggiorno anche gli altri
	if Master {

		var request string = fileName + "|" + fileContent //Build the request in a particular format
		requestJSON, _ := json.Marshal(request)

		for ds := DSList.Front(); ds != nil; ds = ds.Next() {

			response, err := http.Post("http://"+fmt.Sprint(ds.Value)+"/put", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
			if err != nil {
				fmt.Println("An error has occurred trying to estabilish a connection with the replica.")
				fmt.Println(err.Error())
				reportDSCrash(fmt.Sprint(ds.Value)) //CHE VA IMPLEMENTATA PER RIPROVARCI ALMENO 1 VOLTA PRIMA DI TOGLIERE IP
				DSList.Remove(ds)
				continue
			}
			responseFromDS, err := ioutil.ReadAll(response.Body) //Receiving http response
			if err != nil {
				fmt.Println("An error has occurred trying to read the requested file.")
				fmt.Println(err.Error())
				return
			}
			fmt.Println(string(responseFromDS))
		}
	}

	//solo dopo aver aggiornato eventualmente le repliche potra fare
	json.NewEncoder(w).Encode("The file was successfully uploaded.")

}
func del(w http.ResponseWriter, r *http.Request) {
	//Fai il delete, come il put. Poi se master è true procedi con aggiornare anche le repliche
	//Questo è possibile perche se è master ha pure dslist. Cosi posso non implementare 4 funz
	//Aggiorno me stesso
	w.Header().Set("Content-Type", "Application/json")
	fileToRemove := analyzeRequest(r)
	err := os.Remove(fileToRemove) // Remove the file
	if err != nil {
		fmt.Println("An error has occurred trying to delete the file.")
		fmt.Println(err.Error())
		return
	}

	//Se sono master aggiorno anche gli altri
	if Master {

		var request string = fileToRemove //Build the request in a particular format
		requestJSON, _ := json.Marshal(request)

		for ds := DSList.Front(); ds != nil; ds = ds.Next() {

			response, err := http.Post("http://"+fmt.Sprint(ds.Value)+"/delete", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
			if err != nil {
				fmt.Println("An error has occurred trying to estabilish a connection with the replica.")
				fmt.Println(err.Error())
				reportDSCrash(fmt.Sprint(ds.Value)) //CHE VA IMPLEMENTATA PER RIPROVARCI ALMENO 1 VOLTA PRIMA DI TOGLIERE IP
				DSList.Remove(ds)
				continue
			}
			responseFromDS, err := ioutil.ReadAll(response.Body) //Receiving http response
			if err != nil {
				fmt.Println("An error has occurred trying to read the requested file.")
				fmt.Println(err.Error())
				return
			}
			fmt.Println(string(responseFromDS))
		}
	}
	json.NewEncoder(w).Encode("The file was successfully removed.")

}
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

func reportDSCrash(dsCrashed string) {
	var request string = dsCrashed //Build the request in a particular format
	requestJSON, _ := json.Marshal(request)
	http.Post("http://"+DiscoveryIP+"/dsCrash", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request

}

func main() {
	register()
	router := mux.NewRouter()
	router.HandleFunc("/put", put).Methods("POST")
	router.HandleFunc("/delete", del).Methods("POST")
	router.HandleFunc("/get/{key}", get).Methods("GET")
	router.HandleFunc("/becomeMaster", becomeMaster).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))
	/*for e := DSList.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}*/ /*CICLA LA LISTA*/
}
func becomeMaster(w http.ResponseWriter, r *http.Request) {
	Master = true
}
func register() {
	requestJSON, _ := json.Marshal("datastore")
	response, err := http.Post("http://"+DiscoveryIP+":8000/register", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Se fallisce riprova ogni 3 secondi
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node.")
		fmt.Println(err.Error())
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8000/register", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body) //Receiving http response
	if strings.Contains(string(responseFromDiscovery), "master") {
		Master = true
		acquireDSList(string(responseFromDiscovery[0 : len(string(responseFromDiscovery))-6]))
		return
	}
}
func acquireDSList(dslist string) {
	var lastindex = 1 //per via delle doppie virgolette iniziali
	for pos, char := range dslist {
		if char == 124 { //quindi se il carattere letto è |
			DSList.PushBack(dslist[lastindex:pos])
			lastindex = pos + 1
		}
	}
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
