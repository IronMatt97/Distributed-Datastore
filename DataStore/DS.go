package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux" //Libreria aggiuntiva presa da github che permette di utilizzare facilmente un servizio di listen and serve su una porta
)

var DiscoveryIP = "172.17.0.2"       //Indirizzo del nodo discovery
var Master bool = false              //Variabile che viene impostata a true solo dal datastore Master
var DSList []string                  //Lista dei datastore mantenuta dal Master
var MASTERip string = ""             //Ip del master, utilizzato dalle repliche che si connettono al sistema per allinearsi
var mutex sync.Mutex                 //Mutex per agire sulle strutture dati condivise tra le diverse goroutine
var latency = 500 * time.Millisecond //Variabile per la simulazione della latenza

func main() {
	register()
	router := mux.NewRouter()
	router.HandleFunc("/put", put).Methods("POST")
	router.HandleFunc("/del", del).Methods("POST")
	router.HandleFunc("/get/{key}", get).Methods("GET")
	router.HandleFunc("/getData", alignNewReplica).Methods("GET")
	router.HandleFunc("/becomeMaster", becomeMaster).Methods("POST")
	router.HandleFunc("/addDs", addDs).Methods("POST")
	router.HandleFunc("/removeDs", removeDs).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

//Use case: il client vuole effettuare una put
func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	time.Sleep(latency) //Latenza simulata per contattare il DS Master
	receivedRequest := analyzeRequest(r)
	var info []string = strings.Split(receivedRequest, "|") //Acquisisci cosa bisogna salvare dalla richiesta dell'API
	var fileName string = info[0]
	var fileContent string = info[1]
	fmt.Println("A put operation has been called. File save request is  '" + fileName + "' = '" + fileContent + "'")
	mutex.Lock() //Inizia ad effettuare operazioni che devono essere mutualmente esclusive
	if _, err := os.Stat(fileName); err == nil {
		json.NewEncoder(w).Encode("The requested file already exists.")
		mutex.Unlock()
		time.Sleep(latency)
		return
	}
	err := ioutil.WriteFile(fileName, []byte(fileContent), 0777) //Salva il file solo se non c'è già
	mutex.Unlock()
	if err != nil {
		fmt.Println("An error has occurred trying to write the file. ")
		json.NewEncoder(w).Encode("An error has occurred trying to write the file.")
		time.Sleep(latency)
		return
	}

	//Il master deve aggiornare anche le repliche
	if Master {
		var request string = fileName + "|" + fileContent
		requestJSON, _ := json.Marshal(request)
		fmt.Println("I am the Datastore Master, so I am going to update with the request: " + request + " all the replicas:")
		for _, ds := range DSList {
			go contactReplicas(ds, requestJSON, "put") // Invio parallelo degli aggiornamenti put alle repliche
		}
	}
	time.Sleep(latency)
	json.NewEncoder(w).Encode("The file was successfully uploaded.")
}

//Use case: il client ha richiesto una operazione di delete
func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	time.Sleep(latency) //Simulazione della latenza
	fileToRemove := analyzeRequest(r)
	fmt.Println("A delete operation has been called: the requested file to remove is '" + fileToRemove + "'")
	mutex.Lock()
	err := os.Remove(fileToRemove) // Elimina il file
	mutex.Unlock()
	if err != nil {
		fmt.Println("An error has occurred trying to delete the file.")
		json.NewEncoder(w).Encode(string("The file you requested could not be removed."))
		time.Sleep(latency)
		return
	}
	//Il master deve aggiornare anche le repliche
	if Master {
		var request string = fileToRemove
		requestJSON, _ := json.Marshal(request)
		fmt.Println("I am the Datastore Master, so I am going to remove the file '" + request + "' from all the replicas.")
		for _, ds := range DSList {
			go contactReplicas(ds, requestJSON, "delete")
		}
	}
	time.Sleep(latency)
	json.NewEncoder(w).Encode("The file was successfully removed.")
}

//Funzione utile all'implementazione del parallelismo di goroutines per la put
func contactReplicas(ds string, requestJSON []byte, mode string) { //caso put info contiene key:value, caso del contiene ds
	fmt.Println("Updating replica " + ds)
	var response *http.Response
	var err error
	time.Sleep(latency)
	if mode == "put" {
		response, err = http.Post("http://"+ds+":8080/put", "application/json", bytes.NewBuffer(requestJSON))
	}
	if mode == "delete" {
		response, err = http.Post("http://"+ds+":8080/del", "application/json", bytes.NewBuffer(requestJSON))
	}
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the replica.")
		reportDSCrash(ds)
		removeDSFromList(ds)
		return
	}
	responseFromDS, _ := ioutil.ReadAll(response.Body)
	fmt.Println("The replica " + ds + " answered : " + string(responseFromDS))
}

//Use case: il client ha richiesto una get
func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	time.Sleep(latency)   //Simulazione della latenza
	params := mux.Vars(r) //Acquisisci i parametri dalla richiesta
	fmt.Println("get called: I wanna read on myself " + params["key"])
	mutex.Lock()
	data, err := ioutil.ReadFile(params["key"]) //Prova a leggere il file richiesto
	mutex.Unlock()
	if err != nil {
		fmt.Println("An error has occurred reading the file.")
		json.NewEncoder(w).Encode("The requested file does not exists.")
		time.Sleep(latency)
		return
	}
	time.Sleep(latency)
	json.NewEncoder(w).Encode(string(data))
}

//Funzione di registrazione al discovery
func register() {
	requestJSON, _ := json.Marshal("datastore")
	fmt.Println("Datastore correctly initialized. Registering on Discovery node " + DiscoveryIP)
	time.Sleep(latency)
	response, err := http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Se fallisce riprova ogni 3 secondi
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node. Retrying...")
		fmt.Println(err.Error())
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body)
	if strings.Contains(string(responseFromDiscovery), "master") { //Il discovery potrebbe rispondere che ora la replica appena connessa è il master
		becomeMaster(nil, nil)
		acquireDSList(string(responseFromDiscovery[1 : len(string(responseFromDiscovery))-6]))
		return
	}
	//Se si arriva fino a qui, significa che si tratta della registrazione di una replica
	resp := cleanResponse(responseFromDiscovery)
	resp = strings.ReplaceAll(resp, "\"", "")
	resp = strings.ReplaceAll(resp, "\\", "")
	resp = strings.ReplaceAll(resp, "\n", "")
	MASTERip = resp
	getDataUntilNow() //La replica appena connessa richiede al master tutti i file per reallinearsi
}

//Funzione di utility per la rimozione di un Datastore dalla lista
func removeDSFromList(dsToRm string) {
	if len(DSList) > 0 {
		var temp []string
		for _, s := range DSList {
			if s != dsToRm {
				temp = append(temp, s)
			}
		}
		mutex.Lock()
		DSList = temp
		mutex.Unlock()
	}
}

//Funzione che contatta il Discovery per avvisarlo del crash di un Datastore replica
func reportDSCrash(dsCrashed string) {
	var request string = dsCrashed
	requestJSON, _ := json.Marshal(request)
	fmt.Println("The datastore " + dsCrashed + " has crashed: reporting to Discovery...")
	time.Sleep(latency)
	_, err := http.Post("http://"+DiscoveryIP+":8080/dsCrash", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil {
		fmt.Println("Looks like the Discovery has crashed. Waitng for it to restart...")
		time.Sleep(3 * time.Second)
		_, err = http.Post("http://"+DiscoveryIP+":8080/dsCrash", "application/json", bytes.NewBuffer(requestJSON))
	}
}

//Funzione per la rimozione di un datastore dalla lista, chiamabile dal Discovery
func removeDs(w http.ResponseWriter, r *http.Request) {
	req := analyzeRequest(r)
	req = strings.ReplaceAll(req, "\"", "")
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
	fmt.Println("A Datastore replica was removed: the list is now")
	fmt.Println(DSList)
}

//Funzione per aggiungere un Datastore alla lista, chiamabile dal Discovery
func addDs(w http.ResponseWriter, r *http.Request) {
	req := analyzeRequest(r)
	req = strings.ReplaceAll(req, "\"", "")
	fmt.Println("DS prima dell'addDS")
	fmt.Println(DSList)
	if !isInlist(req, DSList) {
		mutex.Lock()
		DSList = append(DSList, req)
		mutex.Unlock()
	}
	fmt.Println("A Datastore replica joined the system: the list is now")
	fmt.Println(DSList)
}

//Funzione chiamabile dal Discovery per l'elezione de Master
func becomeMaster(w http.ResponseWriter, r *http.Request) {
	fmt.Println("|-----------------|")
	fmt.Println("| I AM THE MASTER |")
	fmt.Println("|-----------------|")
	Master = true
	//Se entra in questo if significa che è stato chiamato dal discovery dopo un crash di un altro Master, quindi deve aggiornarsi.
	if r != nil {
		req := analyzeRequest(r)
		acquireDSList(req)
	}
}

//Funzione di utility per il controllo di appartenenza alla lista di una stringa
func isInlist(e string, l []string) bool {
	for _, elem := range l {
		if strings.Compare(e, elem) == 0 {
			return true
		}
	}
	return false
}

//Funzione di utility per la pulizia delle stringhe
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

//Funzione usata dalle repliche appena unite al sistema, che resettano lo stato dei file precedente prima di riacquisirlo dal Master
func flushLocalfiles() {
	mutex.Lock()
	f, err := os.Open(".") //Leggi localmente i vecchi file
	mutex.Unlock()
	if err != nil {
		log.Fatal(err)
	}

	files, err := f.Readdir(-1)

	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.Compare(file.Name(), "DS") != 0 && strings.Compare(file.Name(), "DS.go") != 0 {
			mutex.Lock()
			err := os.Remove(file.Name()) // Rimuovi tutti i file tranne l'eseguibile
			mutex.Unlock()
			if err != nil {
				fmt.Println("An error has occurred trying to delete the file " + file.Name())
				return
			}
		}
	}
}

//Funzione di reallineamento della replica appena unita al sistema
func getDataUntilNow() {
	flushLocalfiles()
	time.Sleep(latency)
	response, err := http.Get("http://" + MASTERip + ":8080/getData")
	if err != nil {
		fmt.Println("An error has occurred trying to acquire data from the Datasore Master")
		return
	}
	data, _ := ioutil.ReadAll(response.Body)
	dataList := string(data)
	var lastindex = 1 //per via delle doppie virgolette iniziali
	var mode int = 0  //mode 0 rappresenta l'acquisizione di filename, mode 1 l'acquisizione del fileContent
	var fileName string
	var fileContent string
	for pos, char := range dataList {
		if char == 124 { //quindi se il carattere letto è |
			if mode == 0 {
				fileName = dataList[lastindex:pos]
				mode = 1
				lastindex = pos + 1
				continue
			}
			if mode == 1 {
				fileContent = dataList[lastindex:pos]
				mode = 0
				lastindex = pos + 1
				//Procedo col salvare il file
				temp := len(fileContent) - 2
				if temp < 1 {
					temp = 1
				}
				fmt.Println("File acquired from the Master -> " + fileName + " : " + fileContent)
				if _, err := os.Stat(fileName); err == nil {
					fmt.Println("The file is already here.") //File già presente
					continue
				}
				mutex.Lock()
				err := ioutil.WriteFile(fileName, []byte(fileContent), 0777)
				mutex.Unlock()
				if err != nil {
					fmt.Println("An error has occurred trying to write the file. ")
					continue
				}
			}
		}
	}

}

//Funzione che le repliche chiamano sul Master per farsi consegnare la lista dei file già presente per reallinearsi
func alignNewReplica(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	dataList := prepareDataList()
	json.NewEncoder(w).Encode(dataList)
}

//Funzione di utility per la costruzione della risposta del DSMaster ai DS replica quando deve consegnargli la lista di file già presenti
func prepareDataList() string {
	mutex.Lock()
	f, err := os.Open(".") //Legge i file localmente
	mutex.Unlock()
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	var list string
	for _, file := range files {
		if strings.Compare(file.Name(), "DS") != 0 && strings.Compare(file.Name(), "DS.go") != 0 {
			mutex.Lock()
			fileContent, _ := ioutil.ReadFile(file.Name())
			mutex.Unlock()
			list = list + file.Name() + "|" + string(fileContent) + "|"
		}
	}
	return list
}

//Funzione di utility che crea la lista dei Datastore a partire dalla risposta nel formato differente di un altro nodo
func acquireDSList(dslist string) {
	var lastindex = 0 //per via delle doppie virgolette iniziali
	for pos, char := range dslist {
		if char == 124 { //quindi se il carattere letto è |
			mutex.Lock()
			DSList = append(DSList, dslist[lastindex:pos])
			mutex.Unlock()
			lastindex = pos + 1
		}
	}
}

//Funzione di utility per convertire la richiesta in un formato leggibile
func analyzeRequest(r *http.Request) string {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("An error has occurred trying to read api's request. ")
		fmt.Println(err.Error())
		return ""
	}
	var receivedRequest string
	err = json.Unmarshal([]byte(requestBody), &receivedRequest)
	if err != nil {
		fmt.Println("Error unmarshaling api's request.")
		fmt.Println(err.Error())
		return ""
	}
	return receivedRequest
}
