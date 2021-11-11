package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux" //Libreria aggiuntiva presa da github che permette di utilizzare facilmente un servizio di listen and serve su una porta
)

var MasterIP string = "" //Indirizzo del Datastore Master
var DSlist []string      //Lista dei Datastore
var restAPIlist []string //Lista delle Api
var mutex sync.Mutex     //Mutex per agire su strutture di dati condivise

//Funzione per scegliere una API da assegnare al client quando si unisce al sistema
func chooseAPI() string {
	apiNum := len(restAPIlist)
	if apiNum == 0 {
		return "noapi"
	}
	n := rand.Intn(apiNum)
	return restAPIlist[n]
}

//Funzione per la rimozione di un datastore dalla lista in seguito ad un crash ad esempio
func removeDSFromList(dsToRemove string) {
	if len(DSlist) > 0 {
		var temp []string
		for _, s := range DSlist {
			if s != dsToRemove {
				temp = append(temp, s)
			}
		}
		mutex.Lock()
		DSlist = temp
		mutex.Unlock()
	}
}

//Funzione per la rimozione di un datastore per via di un crash, chiamabile dagli altri nodi
func dsCrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	dsToRemove := analyzeRequest(r)
	fmt.Println(" I was informed about the datastore " + dsToRemove + "'s crash.")
	removeDSFromList(dsToRemove)
	mutex.Lock()
	os.Remove("DS-" + dsToRemove)
	mutex.Unlock()
	fmt.Println("The datastore was removed: the list is now: ")
	fmt.Println(DSlist)
	fmt.Println("The master is currently the Datastore " + MasterIP)
	requestJSON, _ := json.Marshal(dsToRemove)
	for _, api := range restAPIlist {
		http.Post("http://"+api+":8080/removeDs", "application/json", bytes.NewBuffer(requestJSON))
		fmt.Println("Informing API " + api + " about the Datastore's crash")
	}
	if dsToRemove == MasterIP { //Se a crashare è stato il master, bisogna eleggerne uno nuovo
		fmt.Println("The master is the one who crashed.")
		electNewMaster()
		if MasterIP == "" {
			fmt.Println("There is not a Datastore Master at the moment.")
			return
		}
		requestJSON, _ := json.Marshal(buildDSList())
		http.Post("http://"+MasterIP+":8080/becomeMaster", "application/json", bytes.NewBuffer(requestJSON)) //Avvisa il nuovo master che ora è master
		fmt.Println("The new master " + MasterIP + " was just informed that he is the new master now.")
		requestJSON, _ = json.Marshal(MasterIP) //Ora bisogna avvisare le API del nuovo master
		for _, api := range restAPIlist {
			fmt.Println("The new master is " + MasterIP + " and I am telling it to API " + api)
			_, err := http.Post("http://"+api+":8080/changeMaster", "application/json", bytes.NewBuffer(requestJSON))
			if err != nil {
				fmt.Println("The API " + api + "has crashed. Removing it from the list...")
				removeAPIFromList(api)
				mutex.Lock()
				os.Remove("API-" + api)
				mutex.Unlock()
				fmt.Println("API removed: the list is now ")
				fmt.Println(restAPIlist)
			}
		}
		return
	} //Se arriva fino a qui, non era crashato il master ma un semplice Datastore, quindi procede col dirlo al master
	fmt.Println("Informing Datastore Master " + MasterIP + " about the Datastore's crash")
	http.Post("http://"+MasterIP+":8080/removeDs", "application/json", bytes.NewBuffer(requestJSON))
}

//Funzione dedita all'elezione di un nuovo master in seguito al crash dell'attuale, chiamabile dall'esterno
func dsMasterCrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	fmt.Println("I was informed about the Master Datastore's crash.")
	removeDSFromList(MasterIP)
	mutex.Lock()
	os.Remove("DS-" + MasterIP)
	mutex.Unlock()
	electNewMaster()
	if strings.Compare(MasterIP, "") == 0 {
		fmt.Println("There is not a Master to elect right now.")
		return
	}
	requestJSON, _ := json.Marshal(buildDSList())
	fmt.Println("The new master " + MasterIP + " was just informed that he is the new master now.")
	_, err := http.Post("http://"+MasterIP+":8080/becomeMaster", "application/json", bytes.NewBuffer(requestJSON)) //Avvisa il nuovo master che ora è master
	for err != nil {
		fmt.Println("Looks like the new Master has crashed as well.")
		removeDSFromList(MasterIP)
		mutex.Lock()
		os.Remove("DS-" + MasterIP)
		mutex.Unlock()
		electNewMaster()
		if strings.Compare(MasterIP, "") == 0 {
			fmt.Println("There is not a Master to elect right now.")
			return
		}
		requestJSON, _ := json.Marshal(buildDSList())
		fmt.Println("The new master " + MasterIP + " was just informed that he is the new master now.")
		_, err = http.Post("http://"+MasterIP+":8080/becomeMaster", "application/json", bytes.NewBuffer(requestJSON)) //Avvisa il nuovo master che ora è master
	}
	requestJSON, _ = json.Marshal(MasterIP) //Le api vengono informate del nuovo Master
	for _, api := range restAPIlist {
		fmt.Println("The new master is " + MasterIP + " and I am telling it to API " + api)
		_, err := http.Post("http://"+api+":8080/changeMaster", "application/json", bytes.NewBuffer(requestJSON))
		if err != nil {
			fmt.Println("The API " + api + "has crashed. Removing it from the list...")
			removeAPIFromList(api)
			mutex.Lock()
			os.Remove("API-" + api)
			mutex.Unlock()
			fmt.Println("API removed: the list is now ")
			fmt.Println(restAPIlist)
		}
	}
}

//Funzione contenente la logica di elezione del nuovo Master
func electNewMaster() {
	fmt.Println("---------------------------------------------------")
	fmt.Println("Master election process initialized.")
	fmt.Println("Master ip before election :" + MasterIP)
	MasterIP = ""
	fmt.Println("Candidate DS list:")
	fmt.Println(DSlist)
	if !dsListEmpty() {
		MasterIP = DSlist[0]
	}
	fmt.Println("Elected Master ip :" + MasterIP)
	fmt.Println("----------------------------------------------------")
}

//Funzione fondamentale del discovery che serve a registrare qualsiasi altro nodo voglia connettersi al sistema
func registerNewNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	receivedRequest := analyzeRequest(r)
	response := ""
	fmt.Println("A new node is trying to join the system: the node is a " + receivedRequest)
	fmt.Println("Current system components:")
	fmt.Println(DSlist)
	fmt.Println(restAPIlist)
	if strings.Compare(receivedRequest, "datastore") == 0 { //Caso di registrazione di un Datastore
		fmt.Println("Datastore registration initialized.")
		dsIP := acquireIP(r.RemoteAddr, "datastore") //Aggiungi alla lista di ip e restituiscilo
		mutex.Lock()
		err := ioutil.WriteFile("DS-"+dsIP, []byte(dsIP), 0777) //Salva l'ip localmente, per aumentare la tolleranza ai guasti
		mutex.Unlock()
		if err != nil {
			fmt.Println("An error has occurred trying to register the datastore. ")
			return
		}
		if MasterIP == "" || MasterIP == dsIP { //Se non è presente un master, o il datastore che sta facendo join è il vecchio master restartato
			MasterIP = dsIP
			response = buildDSList() + "master"
			fmt.Println("The new Datastore Master is " + MasterIP + ". All APis will be informed.")
			requestJSON, _ := json.Marshal(MasterIP)
			for _, api := range restAPIlist { //Appena eletto un nuovo master dillo a tutti
				fmt.Println("The new master is " + MasterIP + " and I am telling it to API " + api)
				_, err := http.Post("http://"+api+":8080/changeMaster", "application/json", bytes.NewBuffer(requestJSON))
				if err != nil {
					fmt.Println("The API " + api + "has crashed. Removing it from the list...")
					removeAPIFromList(api)
					mutex.Lock()
					os.Remove("API-" + api)
					mutex.Unlock()
					fmt.Println("API removed: the list is now ")
					fmt.Println(restAPIlist)
				}
			}
		} else { //Caso in cui è semplicmente un datastore ulteriore a volersi unire
			requestJSON, _ := json.Marshal(dsIP)
			fmt.Println("Informing the Master about the new replica...")
			http.Post("http://"+MasterIP+":8080/addDs", "application/json", bytes.NewBuffer(requestJSON)) //avvisa il master che c'è un nuovo DS
			json.NewEncoder(w).Encode(MasterIP)                                                           //Avvisa la replica dell'indirizzo del master
			fmt.Println("Informing the APis about the new replica...")
			for _, api := range restAPIlist {
				fmt.Println("The new ds is " + dsIP + " and I am telling it to API " + api)
				_, err := http.Post("http://"+api+":8080/addDs", "application/json", bytes.NewBuffer(requestJSON))
				if err != nil {
					fmt.Println("The API " + api + "has crashed. Removing it from the list...")
					removeAPIFromList(api)
					mutex.Lock()
					os.Remove("API-" + api)
					mutex.Unlock()
					fmt.Println("API removed: the list is now ")
					fmt.Println(restAPIlist)
				}
			}
			fmt.Println("DS list is:")
			fmt.Println(DSlist)
			return
		}
	}
	if strings.Compare(receivedRequest, "restAPI") == 0 { //Caso in cui è una API a registrarsi
		fmt.Println("An API has joined the system")
		restAPI_IP := acquireIP(r.RemoteAddr, "restAPI") //Aggiungi alla lista di ip e restituiscilo
		mutex.Lock()
		err := ioutil.WriteFile("API-"+restAPI_IP, []byte(restAPI_IP), 0777) //Salvala localmente per tolleranza ai guasti
		mutex.Unlock()
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
			mutex.Lock()
			restAPIlist = append(restAPIlist, restAPI_IP)
			mutex.Unlock()
		}

	}
	if strings.Compare(receivedRequest, "client") == 0 { //Caso in cui è il client a connettersi al sistema
		//registra un nuovo client, semplicemente informandolo di una api da usare
		fmt.Println("A client has joined the system. Choosing an API to return...")
		api := chooseAPI() //restituzione dell'api secondo logica round robin
		response = api
	}
	json.NewEncoder(w).Encode(response)
	fmt.Println("I answered: " + response)
	fmt.Println("API list: ")
	fmt.Println(restAPIlist)
	fmt.Println("Datastore list: ")
	fmt.Println(DSlist)

}

//Funzione di utility per la rimozione di stringhe da liste
func removeAPIFromList(apiToRemove string) {
	if len(restAPIlist) > 0 {
		var temp []string
		for _, api := range restAPIlist {
			if api != apiToRemove {
				temp = append(temp, api)
			}
			mutex.Lock()
			restAPIlist = temp
			mutex.Unlock()
		}
	}
}

//Funzione chiamabile dal client per informare il Discovery del crash di una API
func apicrash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	apiToRemove := analyzeRequest(r)
	removeAPIFromList(apiToRemove)
	mutex.Lock()
	os.Remove("API-" + apiToRemove)
	mutex.Unlock()
	fmt.Println("An API was removed: the list is now")
	fmt.Println(restAPIlist)
	//consegna nuova api da usare
	api := chooseAPI()
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
	router.HandleFunc("/whoisMaster", whoIsMaster).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))

}

//Funzione di utility chiamabile dall'esterno per consegnare l'indirizzo del master su richiesta
func whoIsMaster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	json.NewEncoder(w).Encode(MasterIP)
	fmt.Println("I was asked about Master's address. I answered " + MasterIP)
}

//Funzione di utility per la costruzione della richiesta a partire dal formato lista
func buildDSList() string {
	l := ""
	for _, ds := range DSlist {
		if ds != MasterIP {

			l = l + ds + "|"
		}
	}
	return l
}

//Funzione di recovery del Discovery; quando si avvia controlla sempre che non ci fosse già qualcuno nel sistema, per recuperare lo stato
func checkForPrevState() {
	mutex.Lock()
	files, err := ioutil.ReadDir(".") //Controlla localmente che non ci fosse già qualcuno
	mutex.Unlock()
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files { //Riacquisisci le liste precedentemente salvate in locale, così da recuperare lo stato del sistema.
		if file.Name() != "Discovery.go" {
			fmt.Println("I found already someone in the system: " + file.Name())
		}
		if strings.Contains(file.Name(), "DS-") {
			mutex.Lock()
			DSlist = append(DSlist, file.Name()[3:])
			mutex.Unlock()
		} else if strings.Contains(file.Name(), "API-") {
			mutex.Lock()
			restAPIlist = append(restAPIlist, file.Name()[4:])
			mutex.Unlock()
		}
	}
	//A questo punto iterroga le API per sapere chi era il master
	if !apiListEmpty() {
		for _, api := range restAPIlist {
			fmt.Println("Asking Master's address to API " + api)
			response, err := http.Post("http://"+api+":8080/whoIsMaster", "application/json", nil)
			if err != nil {
				fmt.Println("Error occurred while asking API " + api + " who is the master. Asking to the next if present.")
				continue
			}
			responseFromAPI, _ := ioutil.ReadAll(response.Body)
			MasterIP = cleanResponse((responseFromAPI))
			MasterIP = MasterIP[1 : len(MasterIP)-2]
		}

		fmt.Println("I reacquired the master, which was " + MasterIP)
	}
}

//Funzione di utility per la pulizia di stringhe
func cleanResponse(r []byte) string {
	str := string(r)
	if strings.Contains(str, "\\") {
		str = strconv.Quote(str)
		str = strings.ReplaceAll(str, "\\", "")
		str = strings.ReplaceAll(str, "\"", "")
		if str[len(str)-1:] == "n" {
			str = str[:len(str)-2]
		}
	}
	return str
}

//Funzione di utility per scoprire se l'apiList è vuota
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

//Funzione di utility per scoprire se la dsList è vuota
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

//Funzione di utility per controllare se una stringa appartiene alla lista
func isInlist(e string, l []string) bool {
	for _, elem := range l {
		if strings.Compare(e, elem) == 0 {
			return true
		}
	}
	return false
}

//Funzione di utility per covertire la sequenza di byte di risposta in una stringa leggibile
func analyzeRequest(r *http.Request) string {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("An error has occurred trying to read client's request. ")
		fmt.Println(err.Error())
		return ""
	}
	var receivedRequest string
	err = json.Unmarshal([]byte(requestBody), &receivedRequest)
	if err != nil {
		fmt.Println("Error unmarshaling client's request.")
		fmt.Println(err.Error())
		return ""
	}
	return receivedRequest
}

//Funzione di utility per l'acquisizione degli indirizzi ip in liste, sia per i ds che per le api
func acquireIP(ip string, mode string) string {

	ip = ip[0:len(ip)-6] + "" //Ritaglia l'ip

	if mode == "datastore" {
		var alreadyExists bool = false
		for _, ds := range DSlist {
			if strings.Compare(ds, ip) == 0 {
				alreadyExists = true
			}
		}
		if !alreadyExists {
			mutex.Lock()
			DSlist = append(DSlist, ip)
			mutex.Unlock()
		}
	} else if mode == "restAPI" {
		var alreadyExists bool = false
		for _, api := range restAPIlist {
			if strings.Compare(api, ip) == 0 {
				alreadyExists = true
			}
		}
		if !alreadyExists {
			mutex.Lock()
			restAPIlist = append(restAPIlist, ip)
			mutex.Unlock()
		}
	}
	return ip
}
