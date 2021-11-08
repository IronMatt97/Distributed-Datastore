package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

var MasterIP string = ""
var DSlist []string
var restAPIlist []string
var apiPointer int = 0
var mutex sync.Mutex

func dsCrash(w http.ResponseWriter, r *http.Request) {
	dsToRemove := analyzeRequest(r)
	for pos, ds := range DSlist {
		if strings.Compare(ds, dsToRemove) == 0 {
			a := DSlist[0:pos]
			for _, s := range DSlist[pos+1:] { //Rimuovilo
				a = append(a, s)
			}
			DSlist = a
		}
	}
	mutex.Lock()
	os.Remove("DS-" + dsToRemove)
	mutex.Unlock()
	fmt.Println("è stato rimosso un ds, ora la lista che mi risulta è ")
	fmt.Println(DSlist)
	//Qui devo controllare che se ora il master non è piu nella lista vuol dire che era crashato il master, dovro dirlo
	//sia alle api che eleggere il nuovo master
	if dsToRemove == MasterIP {
		electNewMaster()
		requestJSON, _ := json.Marshal(buildDSList())
		fmt.Println("La lista che ho comunicato al master " + MasterIP + "è " + buildDSList())
		http.Post("http://"+MasterIP+":8080/becomeMaster", "application/json", bytes.NewBuffer(requestJSON)) //Avvisa il nuovo master che ora è master
		fmt.Println("I just told the new master he is new master now")
		requestJSON, _ = json.Marshal(MasterIP) //Devo mettere in attesa l'api di un nuovo ds
		for _, api := range restAPIlist {
			http.Post("http://"+api+":8080/changeMaster", "application/json", bytes.NewBuffer(requestJSON))
			fmt.Println("The new master is " + MasterIP + " and I am telling it to api :" + api)
		}
	}
}
func dsMasterCrash(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ds master crash was called")
	for pos, ds := range DSlist {
		if strings.Compare(ds, MasterIP) == 0 {
			a := DSlist[0:pos]
			for _, s := range DSlist[pos+1:] { //Rimuovilo
				a = append(a, s)
			}
			DSlist = a
		}
	}
	fmt.Println("In teoria ho rimostto il ds dalla lista, mi risultano i ds: ")
	fmt.Println(DSlist)
	fmt.Println("Sto per rimuovere" + "DS-" + MasterIP)
	mutex.Lock()
	os.Remove("DS-" + MasterIP)
	mutex.Unlock()

	electNewMaster()
	requestJSON, _ := json.Marshal(buildDSList())
	fmt.Println("La lista che ho comunicato al master " + MasterIP + "è " + buildDSList())
	http.Post("http://"+MasterIP+":8080/becomeMaster", "application/json", bytes.NewBuffer(requestJSON)) //Avvisa il nuovo master che ora è master
	fmt.Println("I just told the new master he is new master now")
	requestJSON, _ = json.Marshal(MasterIP) //Devo mettere in attesa l'api di un nuovo ds
	for _, api := range restAPIlist {
		http.Post("http://"+api+":8080/changeMaster", "application/json", bytes.NewBuffer(requestJSON))
		fmt.Println("The new master is " + MasterIP + " and I am telling it to api :" + api)
	}
}
func electNewMaster() {
	fmt.Println("The master has changed, old master was :" + MasterIP)
	MasterIP = ""
	fmt.Println("Mi risulta una lista tra cui scegliere di ds")
	fmt.Println(DSlist)
	if !dsListEmpty() {
		MasterIP = DSlist[0]
	}
	fmt.Println("The master has changed, new master is :" + MasterIP)
}

func registerNewNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")

	receivedRequest := analyzeRequest(r)
	response := ""
	fmt.Println("Somebody registered: its a " + receivedRequest)
	fmt.Println("Per ora le liste sono ds/api")
	fmt.Println(DSlist)
	fmt.Println(restAPIlist)
	if strings.Compare(receivedRequest, "datastore") == 0 {
		fmt.Println("entrato nel caso ds")
		//Register new datastore
		dsIP := acquireIP(r.RemoteAddr, "datastore") //Aggiungi alla lista di ip e restituiscilo
		mutex.Lock()
		err := ioutil.WriteFile("DS-"+dsIP, []byte(dsIP), 0777) //Write the file
		mutex.Unlock()
		if err != nil {
			fmt.Println("An error has occurred trying to register the datastore. ")
			fmt.Println(err.Error())
			return
		}
		if MasterIP == "" || MasterIP == dsIP {
			MasterIP = dsIP
			fmt.Println("sto dichiarando il nuovo master")
			ipList := buildDSList()
			fmt.Println("La lista appena costruita è")
			fmt.Println(ipList)
			response = ipList + "master"
			fmt.Println("Il nuovo master è " + MasterIP + " lo dico alle api presenti ovvero")
			fmt.Println(restAPIlist)
			requestJSON, _ := json.Marshal(MasterIP)
			for _, api := range restAPIlist { //Appena eletto un nuovo master dillo a tutti
				http.Post("http://"+api+":8080/changeMaster", "application/json", bytes.NewBuffer(requestJSON))
				fmt.Println("The new master is " + MasterIP + " and I am telling it to api :" + api)
			}
		} else {
			fmt.Println("I registered a new datastore: " + dsIP + " il master c'era gia quindi ora lo avviso della nuova replica")
			requestJSON, _ := json.Marshal(dsIP)
			fmt.Println("STO AVVISANDO IL MASTER CHE CE UNA NUOVA REPLICA")
			http.Post("http://"+MasterIP+":8080/addDs", "application/json", bytes.NewBuffer(requestJSON)) //avvisa che c'è un nuovo DS
			fmt.Println("Sto rispondendo a " + dsIP + "l'indirizzo del master ovvero " + MasterIP + "per poi ritornare")
			json.NewEncoder(w).Encode(MasterIP)
			fmt.Println("anzi prima avviso pure le api del nuovo ds")
			for _, api := range restAPIlist { //Appena eletto un nuovo ds replca dillo a tutti
				http.Post("http://"+api+":8080/addDs", "application/json", bytes.NewBuffer(requestJSON))
				fmt.Println("The new ds is " + dsIP + " and I am telling it to api :" + api)
			}
			return

		}

	}
	if strings.Compare(receivedRequest, "restAPI") == 0 {
		//Register new restAPI
		fmt.Println("entrato nel caso restapi")
		restAPI_IP := acquireIP(r.RemoteAddr, "restAPI")                     //Aggiungi alla lista di ip e restituiscilo
		err := ioutil.WriteFile("API-"+restAPI_IP, []byte(restAPI_IP), 0777) //Write the file
		if err != nil {
			fmt.Println("An error has occurred trying to register the datastore. ")
			fmt.Println(err.Error())
			return
		}
		response = ""
		for _, ds := range DSlist {
			if strings.Compare(ds, MasterIP) != 0 { //il master va messo alla fine
				response = response + ds + "|"
			}
		}
		response = response + MasterIP + "|"
		fmt.Println("I registered a new restAPI: " + restAPI_IP)
		if !isInlist(restAPI_IP, restAPIlist) {
			restAPIlist = append(restAPIlist, restAPI_IP)
		}

	}
	if strings.Compare(receivedRequest, "client") == 0 {
		//registra un nuovo client, semplicemente informandolo di una api da usare
		fmt.Println("entrato nel caso client")
		if apiPointer > len(restAPIlist)-1 { //bug prevention
			if (len(restAPIlist) - 1) == 0 {

				fmt.Println("non ci sono API disponibili da comunicare al client. ")
				response := "noapi"
				json.NewEncoder(w).Encode(response)
				return
			}
			apiPointer = len(restAPIlist) - 1
		}
		api := restAPIlist[apiPointer]
		apiPointer = (apiPointer + 1) % len(restAPIlist) //restituzione dell'api secondo logica round robin
		response = api
	}
	//Answer requestOK
	json.NewEncoder(w).Encode(response)
	fmt.Println("I answered: " + response)
	fmt.Println("Lista delle API connesse: ")
	fmt.Println(restAPIlist)
	fmt.Println("Lista dei DS connessi: ")
	fmt.Println(DSlist)

}
func apicrash(w http.ResponseWriter, r *http.Request) {
	apiToRemove := analyzeRequest(r)
	for pos, api := range restAPIlist {
		if strings.Compare(api, apiToRemove) == 0 {
			a := restAPIlist[0:pos]
			for _, s := range restAPIlist[pos+1:] { //Rimuovilo
				a = append(a, s)
			}
			restAPIlist = a
		}
	}
	mutex.Lock()
	os.Remove("API-" + apiToRemove)
	mutex.Unlock()
	fmt.Println("è stato rimosso un api, ora la lista che mi risulta è ")
	fmt.Println(restAPIlist)
	//consegna nuova api da usare
	if apiPointer > len(restAPIlist)-1 { //bug prevention
		if (len(restAPIlist) - 1) == 0 {

			fmt.Println("non ci sono API disponibili da comunicare al client. ")
			response := "noapi"
			json.NewEncoder(w).Encode(response)
			return
		}
		apiPointer = len(restAPIlist) - 1
	}
	api := restAPIlist[apiPointer]
	apiPointer = (apiPointer + 1) % len(restAPIlist) //restituzione dell'api secondo logica round robin
	response := api
	json.NewEncoder(w).Encode(response)
}
func main() {
	checkForPrevState()
	router := mux.NewRouter()
	router.HandleFunc("/register", registerNewNode).Methods("POST")
	router.HandleFunc("/dsCrash", dsCrash).Methods("POST")
	router.HandleFunc("/dsMasterCrash", dsMasterCrash).Methods("POST")
	router.HandleFunc("/apicrash", apicrash).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))

}
func buildDSList() string {
	l := ""
	for _, ds := range DSlist {
		if ds != MasterIP {

			l = l + ds + "|"
		}
	}
	return l
}
func checkForPrevState() {
	mutex.Lock()
	files, err := ioutil.ReadDir(".")
	mutex.Unlock()
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Println("I found already someone in the system: " + file.Name())
		if strings.Contains(file.Name(), "DS-") {
			DSlist = append(DSlist, file.Name()[3:])
		} else if strings.Contains(file.Name(), "API-") {
			restAPIlist = append(restAPIlist, file.Name()[4:])
		}
	}

	if !apiListEmpty() {
		//Ora interroga una API alla volta per sapere chi è il master
		for _, api := range restAPIlist {
			fmt.Println("Sto provando a chiedere chi è il master all'api " + "http://" + api + ":8080/whoIsMaster")
			requestJSON, _ := json.Marshal("chi è il master?")
			response, err := http.Post("http://"+api+":8080/whoIsMaster", "application/json", bytes.NewBuffer(requestJSON))
			fmt.Println("in teoria ho mandato la post")
			if err != nil {
				fmt.Println("sembrerebbe che ho trovato un errore")
				continue
			}
			fmt.Println("indirizzo valido trovato aggiorno il master.")
			responseFromAPI, err := ioutil.ReadAll(response.Body)
			MasterIP = string(responseFromAPI)
		}

		fmt.Println("I acquired the master, which was " + MasterIP)
	}
}
func apiListEmpty() bool {
	c := 0
	for range restAPIlist {
		c++
	}
	if c == 0 {
		return true
	} else {
		return false
	}
}
func dsListEmpty() bool {
	c := 0
	for range DSlist {
		c++
	}
	if c == 0 {
		return true
	} else {
		return false
	}
}
func isInlist(e string, l []string) bool {
	for _, elem := range l {
		if strings.Compare(e, elem) == 0 {
			return true
		}
	}
	return false
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
		var alreadyExists bool = false
		for _, ds := range DSlist {
			if strings.Compare(ds, ip) == 0 {
				alreadyExists = true
			}
		}
		if !alreadyExists {
			DSlist = append(DSlist, ip)
		}
	} else if mode == "restAPI" {
		var alreadyExists bool = false
		for _, api := range restAPIlist {
			if strings.Compare(api, ip) == 0 {
				alreadyExists = true
			}
		}
		if !alreadyExists {
			restAPIlist = append(restAPIlist, ip)
		}
	}
	return ip
}
