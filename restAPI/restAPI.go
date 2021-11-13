package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux" //Libreria aggiuntiva presa da github che permette di utilizzare facilmente un servizio di listen and serve su una porta
)

var DiscoveryIP string = "172.17.0.2" //Indirizzo del nodo Discovery
var DSMasterIP string = ""            //Indirizzo del Datastore Master, necessario per le operazioni di put e del
var DSList []string                   //Lista dei Datastore, necessaria per le operazioni di get distribuite
var mutex sync.Mutex                  //Mutex per l'utilizzo di strutture dati condivise tra le goroutine

func main() {
	register()                //Registrazione al Discovery
	router := mux.NewRouter() //Inizializzazione del router
	router.HandleFunc("/get/{key}", get).Methods("GET")
	router.HandleFunc("/put", put).Methods("POST") //handler/endpoint per ogni use case e per altre funzionalità utili
	router.HandleFunc("/del", del).Methods("POST")
	router.HandleFunc("/changeMaster", changeDSMasterOnCrash).Methods("POST")
	router.HandleFunc("/whoIsMaster", whoIsMaster).Methods("POST")
	router.HandleFunc("/addDs", addDs).Methods("POST")
	router.HandleFunc("/removeDs", removeDs).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router)) //Listen and serve delle richieste sulla porta 8080
}

//Use case: il client vuole effettuare una get
func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fmt.Println("A get operation has been called.")
	params := mux.Vars(r)
	ds := chooseDS() //Scegli il datastore tra i conosciuti
	if ds == "" {
		fmt.Println("Looks like there are not Datastores up at the moment. Please retry later.")
		json.NewEncoder(w).Encode("There are not Datastores to connect to. Please retry later.")
		return
	}
	fmt.Println("The get operation will query the file '" + params["key"] + "' on the Datastore " + ds)
	response, err := http.Get("http://" + ds + ":8080/get/" + params["key"])
	for err != nil { //Qualora il datastore fosse crashato nel frattempo
		if ds == DSMasterIP {
			removeDSFromList(DSMasterIP)
			DSMasterIP = ""
			reportDSMasterCrash()
			time.Sleep(3 * time.Second)
			response, err = http.Get("http://" + ds + ":8080/get/" + params["key"])
			continue
		}
		reportDSCrash(ds)
		time.Sleep(3 * time.Second)
		response, err = http.Get("http://" + ds + ":8080/get/" + params["key"])
	}
	responseFromDS, _ := ioutil.ReadAll(response.Body)
	json.NewEncoder(w).Encode(string(responseFromDS)) //Invia al client la risposta del Datastore
}

//Use case: il client vuole effettuare una put
func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	receivedRequest := analyzeRequest(r)
	var info []string = strings.Split(receivedRequest, "|") //Acquisisci il nome ed il contenuto del file dalla richiesta del client
	var fileName string = info[0]
	var fileContent string = info[1]
	var request string = fileName + "|" + fileContent //Ricostruisci la richiesta per il Datastore Master
	for DSMasterIP == "" {                            //Se al momento non c'è un master prova a chiedere a Discovery se ne è arrivato uno, riprovando se non risponde
		fmt.Println("There is not a Datastore Master at the moment. Waiting for a Master to come back ...")
		time.Sleep(3 * time.Second)
		response, _ := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil)
		r, _ := ioutil.ReadAll(response.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1) //Effettua la pulizia della stringa dai caratteri indesiderati
	}
	fmt.Println("A put operation has been called: the Client wants to write " + request + " on the Master Datastore " + DSMasterIP)
	requestJSON, _ := json.Marshal(request)
	response, err := http.Post("http://"+DSMasterIP+":8080/put", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil {
		reportDSMasterCrash() //Se il Master è crashato
		removeDSFromList(DSMasterIP)
		DSMasterIP = ""
		resp, e := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Richiedi chi è il master
		for e != nil {
			fmt.Println("The Discovery node has crashed. Waiting for it to come back...")
			time.Sleep(3 * time.Second)
			resp, e = http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil)
		}
		r, _ := ioutil.ReadAll(resp.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DSMasterIP+":8080/put", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDS, _ := ioutil.ReadAll(response.Body)
	json.NewEncoder(w).Encode(string(responseFromDS))
}

//Use case: il client vuole effettuare una operazione di del
func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fileToRemove := analyzeRequest(r)
	var request string = fileToRemove
	for DSMasterIP == "" { //Fino a quando non torna un Master
		fmt.Println("There is not a Datastore Master at the moment. Waiting for a Master to come back ...")
		time.Sleep(3 * time.Second)
		response, _ := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil)
		r, _ := ioutil.ReadAll(response.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
	}
	fmt.Println("A delete operation has been called: the file to remove is '" + fileToRemove + "', on the Master Datastore " + DSMasterIP)
	requestJSON, _ := json.Marshal(request)
	response, err := http.Post("http://"+DSMasterIP+":8080/del", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil {
		reportDSMasterCrash()
		removeDSFromList(DSMasterIP)
		DSMasterIP = ""
		resp, e := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil)
		for e != nil {
			fmt.Println("Discovery crashato, aspetto che torna")
			time.Sleep(3 * time.Second)
			resp, e = http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil)
		}
		r, _ := ioutil.ReadAll(resp.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DSMasterIP+":8080/del", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDS, _ := ioutil.ReadAll(response.Body)
	json.NewEncoder(w).Encode(string(responseFromDS))
}

//Funzione di registrazione al sistema
func register() {
	fmt.Println("API node initialized: starting the registration process...")
	requestJSON, _ := json.Marshal("restAPI")
	response, err := http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Se fallisce riprova ogni 3 secondi
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node. Retrying...")
		fmt.Println(err.Error())
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body)
	for strings.Compare(string(responseFromDiscovery)[1:len(string(responseFromDiscovery))-2], "|") == 0 {
		fmt.Println("The master is not here yet. I am gonna wait...")
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
		if err != nil {
			fmt.Println("The Discovery node has crashed. Trying to reconnect ...")
			for err != nil {
				response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
			}
		}
		responseFromDiscovery, _ = ioutil.ReadAll(response.Body)
	}
	var dslist string = (string(responseFromDiscovery[0 : len(string(responseFromDiscovery))-2]))
	dslist = strings.ReplaceAll(dslist, "\\", "")
	dslist = strings.ReplaceAll(dslist, "n", "") //Pulizia dell'output
	dslist = strings.ReplaceAll(dslist, "\"", "")
	acquireDSList(dslist) //vengono restituiti tutti i ds con il master alla fine
	DSMasterIP = DSList[len(DSList)-1]
	fmt.Println("Registration process complete: the master is " + DSMasterIP)
	fmt.Println("The Datastore list is: ")
	fmt.Println(DSList)
}

//Funzione di utility che viene chiamata dal Discovery per informare questo nodo API dell'elezione di un nuovo DS Master
func changeDSMasterOnCrash(w http.ResponseWriter, r *http.Request) {
	DSMasterIP = analyzeRequest(r)
	fmt.Println("A new Datastore Master has been elected, the new Master is " + DSMasterIP)
}

//Funzione di recovery chiamata dal Datastore quando riparte per riaggiornarsi sullo stato del sistema
func whoIsMaster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fmt.Println("Il discovery mi ha chiesto chi è il master, e per me è " + DSMasterIP)
	json.NewEncoder(w).Encode(DSMasterIP)
}

//Funzione chiamata dal Discovery per aggiungere un nuovo Datastore alla lista
func addDs(w http.ResponseWriter, r *http.Request) {
	req := analyzeRequest(r)
	if !isInlist(req, DSList) {
		mutex.Lock()
		DSList = append(DSList, req)
		mutex.Unlock()
	}
	fmt.Println("A new Datastore joined the system. The list is now: ")
	fmt.Println(DSList)
}

//Funzione di utility per rimuovere un Datastore dalla lista
func removeDs(w http.ResponseWriter, r *http.Request) {

	req := analyzeRequest(r)
	if !isInlist(req, DSList) {
		return
	}
	//rimuovi il ds dalla lista
	var t []string
	for _, ds := range DSList {
		if ds != req {
			t = append(t, ds)
		}
	}
	mutex.Lock()
	DSList = t
	mutex.Unlock()
	fmt.Println("A Datastore has been removed. The resulting list is now: ")
	fmt.Println(DSList)
}

//Funzione per la scelta di un DS tra quelli presenti al fine di fare la get
func chooseDS() string {
	dsNum := len(DSList)
	if dsNum == 0 {
		response, _ := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Se risultano 0 DS, controlla se è stato rieletto un master
		r, _ := ioutil.ReadAll(response.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
		if DSMasterIP != "" {
			mutex.Lock()
			DSList = append(DSList, DSMasterIP) //Se è tornato un master, riaggiungilo alla lista dei datastore
			mutex.Unlock()
		}
		return DSMasterIP
	}
	n := rand.Intn(dsNum)
	return DSList[n] //Ritorna il DS scelto tra quelli presenti
}

//Funzione che comunica al nodo discovery il crash del Datastore Master
func reportDSMasterCrash() {
	fmt.Println("Looks like the Master Datastore has crashed: this will be reported to the Discovery node.")
	var request string = DSMasterIP
	requestJSON, _ := json.Marshal(request)
	_, err := http.Post("http://"+DiscoveryIP+":8080/dsMasterCrash", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Riprova ogni 3 secondi se non riesce a contattare il Discovery
		fmt.Println("The Discovery node is down at the moment. Waiting for it to restart...")
		time.Sleep(3 * time.Second)
		_, err = http.Post("http://"+DiscoveryIP+":8080/dsMasterCrash", "application/json", bytes.NewBuffer(requestJSON))
	}
	fmt.Println("The Discovery node has been correctly informed of the Master Datastore crash.")
}

//Funzione che comunica al nodo discovery il crash di un Datastore
func reportDSCrash(dsCrashed string) {
	var request string = dsCrashed
	requestJSON, _ := json.Marshal(request)
	fmt.Println("The datastore " + dsCrashed + " has crashed: this will be reported to the Discovery node.")
	_, err := http.Post("http://"+DiscoveryIP+":8080/dsCrash", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil {
		fmt.Println("The Discovery node is down at the moment. Waiting for it to restart...")
		time.Sleep(3 * time.Second)
		_, err = http.Post("http://"+DiscoveryIP+":8080/dsCrash", "application/json", bytes.NewBuffer(requestJSON))
	}
	removeDSFromList(dsCrashed) //Rimuovi il Datasore dalla lista conosciuta
	fmt.Println("The crashed Datastore has been removed from the list. The list is now:")
	fmt.Println(DSList)
}

//Funzione di utility che rimuove un Datastore dalla lista
func removeDSFromList(dsToRemove string) {
	if len(DSList) > 0 {
		var t []string
		for _, ds := range DSList {
			if ds != dsToRemove {
				t = append(t, ds)
			}
		}
		mutex.Lock()
		DSList = t     //La lista è acceduta dalle diverse goroutine dei diversi client che potrebbero contattare la stessa api
		mutex.Unlock() //Ha senso sincronizzare le letture e le scritture sulla struttura di dati condivisa
	}
}

//Funzione di utility per vedere se una stringa appartiene alla lista
func isInlist(e string, l []string) bool {
	for _, elem := range l {
		if strings.Compare(e, elem) == 0 {
			return true
		}
	}
	return false
}

//Funzione di utility per estrapolare la lista dei DS dalla risposta del discovery
func acquireDSList(dslist string) {
	var lastindex int = 0
	for pos, char := range dslist {
		if char == 124 { //quindi se il carattere letto è |
			if !isInlist(dslist[lastindex:pos], DSList) {
				mutex.Lock()
				DSList = append(DSList, dslist[lastindex:pos]) //Separa la stringa ed appendila alla lista
				mutex.Unlock()
			}
			lastindex = pos + 1
		}
	}
}

//Funzione di utility per l'estrapolazione della stringa dalla sequenza di bytes ricevuti
func analyzeRequest(r *http.Request) string {
	requestBody, err := ioutil.ReadAll(r.Body) //Leggi la richiesta
	if err != nil {
		fmt.Println("An error has occurred trying to read client's request. ")
		fmt.Println(err.Error())
		return ""
	}
	var receivedRequest string
	err = json.Unmarshal([]byte(requestBody), &receivedRequest) //Unmarshal della richiesta
	if err != nil {
		fmt.Println("Error unmarshaling client's request.")
		fmt.Println(err.Error())
		return ""
	}
	return receivedRequest
}
