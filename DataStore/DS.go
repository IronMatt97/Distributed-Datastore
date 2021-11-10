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

	"github.com/gorilla/mux"
)

var DiscoveryIP = "172.17.0.2"
var Master bool = false
var DSList []string
var MASTERip string = ""
var mutex sync.Mutex

func put(w http.ResponseWriter, r *http.Request) {
	//provaMutex.Lock()
	//Aggiorno me stesso
	w.Header().Set("Content-Type", "Application/json")
	time.Sleep(2 * time.Second)
	receivedRequest := analyzeRequest(r)
	var info []string = strings.Split(receivedRequest, "|") //Acquire file name and content from client's request
	var fileName string = info[0]
	var fileContent string = info[1]
	fmt.Println("put called: I wanna write on myself " + fileName + " : " + fileContent)

	mutex.Lock()
	if _, err := os.Stat(fileName); err == nil {
		json.NewEncoder(w).Encode("The file you requested already exists.") //Return error if file already exists
		mutex.Unlock()
		return
	}
	err := ioutil.WriteFile(fileName, []byte(fileContent), 0777) //Write the file
	mutex.Unlock()
	if err != nil {
		fmt.Println("An error has occurred trying to write the file. ")
		fmt.Println(err.Error())
		return
	}

	//Se sono master aggiorno anche gli altri
	if Master {

		var request string = fileName + "|" + fileContent //Build the request in a particular format
		requestJSON, _ := json.Marshal(request)
		fmt.Println("I am master, I am going to update with " + request + " replicas:")
		for pos, ds := range DSList {
			fmt.Println("updating" + ds)
			response, err := http.Post("http://"+ds+":8080/put", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
			if err != nil {
				fmt.Println("An error has occurred trying to estabilish a connection with the replica.")
				fmt.Println(err.Error())
				reportDSCrash(ds) //CHE VA IMPLEMENTATA PER RIPROVARCI ALMENO 1 VOLTA PRIMA DI TOGLIERE IP
				if len(DSList) > 0 {
					a := DSList[0:pos]
					for _, s := range DSList[pos+1:] { //Rimuovilo
						a = append(a, s)
					}
					DSList = a
				}
				continue
			}
			responseFromDS, err := ioutil.ReadAll(response.Body) //Receiving http response
			if err != nil {
				fmt.Println("An error has occurred trying to acquire replica response.")
				fmt.Println(err.Error())
				return
			}
			fmt.Println("replica " + ds + " answer to me: " + string(responseFromDS))
		}
	}

	//solo dopo aver aggiornato eventualmente le repliche potra fare
	json.NewEncoder(w).Encode("The file was successfully uploaded.")
	//provaMutex.Unlock()
}
func del(w http.ResponseWriter, r *http.Request) {
	//provaMutex.Lock()
	//Fai il delete, come il put. Poi se master è true procedi con aggiornare anche le repliche
	//Questo è possibile perche se è master ha pure dslist. Cosi posso non implementare 4 funz
	//Aggiorno me stesso
	w.Header().Set("Content-Type", "Application/json")
	time.Sleep(2 * time.Second)
	fileToRemove := analyzeRequest(r)
	fmt.Println("del called: I wanna del on myself " + fileToRemove)
	mutex.Lock()
	err := os.Remove(fileToRemove) // Remove the file
	mutex.Unlock()
	if err != nil {
		fmt.Println("An error has occurred trying to delete the file.")
		fmt.Println(err.Error())
		json.NewEncoder(w).Encode(string("The file you requested could not be removed."))
		return
	}

	//Se sono master aggiorno anche gli altri
	if Master {

		var request string = fileToRemove //Build the request in a particular format
		requestJSON, _ := json.Marshal(request)
		fmt.Println("I am master, I am going to del " + request + "from replicas:")
		for pos, ds := range DSList {
			fmt.Println("deleting" + ds)
			response, err := http.Post("http://"+ds+":8080/del", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
			if err != nil {
				fmt.Println("An error has occurred trying to estabilish a connection with the replica.")
				fmt.Println(err.Error())
				reportDSCrash(ds) //CHE VA IMPLEMENTATA PER RIPROVARCI ALMENO 1 VOLTA PRIMA DI TOGLIERE IP
				if len(DSList) > 0 {
					a := DSList[0:pos]
					for _, s := range DSList[pos+1:] { //Rimuovilo
						a = append(a, s)
					}
					DSList = a
				}
				continue
			}
			responseFromDS, err := ioutil.ReadAll(response.Body) //Receiving http response
			if err != nil {
				fmt.Println("An error has occurred trying to acquire answer from replica.")
				fmt.Println(err.Error())
				return
			}
			fmt.Println("replica " + ds + " answer to me: " + string(responseFromDS))
		}
	}

	json.NewEncoder(w).Encode("The file was successfully removed.")
	//provaMutex.Unlock()
}
func get(w http.ResponseWriter, r *http.Request) {
	//provaMutex.Lock()
	w.Header().Set("Content-Type", "Application/json")
	time.Sleep(1 * time.Second)
	params := mux.Vars(r) //Acquire url params

	fmt.Println("get called: I wanna read on myself " + params["key"])
	mutex.Lock()
	data, err := ioutil.ReadFile(params["key"]) //Try to read the requested file
	mutex.Unlock()
	if err != nil {
		fmt.Println("An error has occurred reading the file.")
		fmt.Println(err.Error())
		json.NewEncoder(w).Encode("An error has occurred reading the file/file does not exists.")
		return
	}
	time.Sleep(1 * time.Second)
	json.NewEncoder(w).Encode(string(data)) //Send the response to the client
	//provaMutex.Unlock()
}

func reportDSCrash(dsCrashed string) {
	var request string = dsCrashed //Build the request in a particular format
	requestJSON, _ := json.Marshal(request)
	fmt.Println("ds crashed, sending this to discovery ")
	_, err := http.Post("http://"+DiscoveryIP+":8080/dsCrash", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	for err != nil {
		fmt.Println("discovery crashed. Waitng for it to restart.")
		time.Sleep(3 * time.Second)
		_, err = http.Post("http://"+DiscoveryIP+":8080/dsCrash", "application/json", bytes.NewBuffer(requestJSON))
	}
}

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
	/*for e := DSList.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}*/ /*CICLA LA LISTA*/
}
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
	DSList = t
	fmt.Println("rimossa replica: ora l'insieme dei ds è")
	fmt.Println(DSList)
}
func addDs(w http.ResponseWriter, r *http.Request) { //Questa funzione deve solo aggiungere al master la replica
	//in un altra funzione bisogna invece allineare la replica, ma è la replica che deve richiederlo.
	req := analyzeRequest(r)

	if !isInlist(req, DSList) {
		DSList = append(DSList, req)
	}
	fmt.Println("Aggiunta nuova replica: ora l'insieme dei ds è")
	fmt.Println(DSList)
}

func becomeMaster(w http.ResponseWriter, r *http.Request) {
	fmt.Println("---------------")
	fmt.Println("I AM MASTER NOW")
	fmt.Println("---------------")
	Master = true

	//Se entra in questo if è stato chiamato dal discovery dopo un crash
	if r != nil {
		req := analyzeRequest(r)
		acquireDSList(req)
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
func register() {
	requestJSON, _ := json.Marshal("datastore")
	fmt.Println("I am trying to register myself on " + DiscoveryIP)
	response, err := http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Se fallisce riprova ogni 3 secondi
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node.")
		fmt.Println(err.Error())
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body) //Receiving http response
	if strings.Contains(string(responseFromDiscovery), "master") {
		becomeMaster(nil, nil)
		acquireDSList(string(responseFromDiscovery[0 : len(string(responseFromDiscovery))-6]))
		return
	}
	//SE SEI ARRIVATO FINO A QUI SEI UNA REPLICA QUNDI DOVRAI AGGIORNARTI DOPO
	resp := cleanResponse(responseFromDiscovery)
	resp = strings.ReplaceAll(resp, "\"", "")
	resp = strings.ReplaceAll(resp, "\\", "")
	resp = strings.ReplaceAll(resp, "\n", "")
	//MASTERip = string(responseFromDiscovery[1 : len(string(responseFromDiscovery))-2])
	MASTERip = resp
	getDataUntilNow()
}
func cleanResponse(r []byte) string {
	str := string(r)
	if strings.Contains(str, "\\") {
		str = strconv.Quote(str)
		str = strings.ReplaceAll(str, "\\", "")
		str = strings.ReplaceAll(str, "\"", "")
		if str[len(str)-1:] == "n" {
			str = str[:len(str)-2] //Cleaning the output
		}
	}
	return str
}
func flushLocalfiles() {
	f, err := os.Open(".")
	if err != nil {
		log.Fatal(err)
	}
	mutex.Lock()
	files, err := f.Readdir(-1)
	mutex.Unlock()
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.Compare(file.Name(), "DS") != 0 && strings.Compare(file.Name(), "DS.go") != 0 {
			mutex.Lock()
			err := os.Remove(file.Name()) // Remove the file
			mutex.Unlock()
			if err != nil {
				fmt.Println("An error has occurred trying to delete the file " + file.Name())
				fmt.Println(err.Error())
				return
			}
		}
	}
}

func getDataUntilNow() {
	flushLocalfiles()
	response, err := http.Get("http://" + MASTERip + ":8080/getData")
	if err != nil {
		fmt.Println("An error has occurred acquiring the data from master")
		fmt.Println(err.Error())
		return
	}
	data, _ := ioutil.ReadAll(response.Body)
	dataList := string(data)
	var lastindex = 1 //per via delle doppie virgolette iniziali
	var mode int = 0
	var fileName string
	var fileContent string
	for pos, char := range dataList {
		//fmt.Println("sto ciclando, ho per le mani il char , sto in mode " + fmt.Sprintln(mode))
		fmt.Println(char)
		//fmt.Println("controllo che sia 124")
		if char == 124 { //quindi se il carattere letto è |
			//fmt.Println("trovato un carattere 124")
			if mode == 0 {
				//fmt.Println("entrato in if 0 , acquisisco da lastindex a pos ovvero da " + fmt.Sprint(lastindex) + " a " + fmt.Sprint(pos))
				fileName = dataList[lastindex:pos]
				mode = 1
				lastindex = pos + 1
				continue
			}
			if mode == 1 {
				//fmt.Println("entrato in if 1 , acquisisco da lastindex a pos ovvero da " + fmt.Sprint(lastindex) + " a " + fmt.Sprint(pos))
				fileContent = dataList[lastindex:pos]
				mode = 0
				lastindex = pos + 1

				//Procedo col salvare il file
				temp := len(fileContent) - 2
				if temp < 1 {
					temp = 1
				}
				fmt.Println("ho acquisito nome e testo, I wanna write on myself " + fileName + " : " + fileContent)
				if _, err := os.Stat(fileName); err == nil {
					fmt.Println("The file you requested already exists.") //Return error if file already exists
					return
				}

				err := ioutil.WriteFile(fileName, []byte(fileContent), 0777) //Write the file
				if err != nil {
					fmt.Println("An error has occurred trying to write the file. ")
					fmt.Println(err.Error())
					return
				}
			}
		}
	}

}
func alignNewReplica(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	dataList := prepareDataList()
	json.NewEncoder(w).Encode(dataList) //Send the response to the client
}

func prepareDataList() string {
	f, err := os.Open(".")
	if err != nil {
		log.Fatal(err)
	}
	mutex.Lock()
	files, err := f.Readdir(-1)
	mutex.Unlock()
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
func acquireDSList(dslist string) {
	strings.ReplaceAll(dslist, "\"", "")
	var lastindex = 0 //per via delle doppie virgolette iniziali
	for pos, char := range dslist {
		if char == 124 { //quindi se il carattere letto è |
			DSList = append(DSList, dslist[lastindex:pos])
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
