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
	"time"

	"github.com/gorilla/mux"
)

var DiscoveryIP string = "172.17.0.2"
var DSMasterIP string = ""
var DSList []string

func chooseDS() string {
	dsNum := len(DSList)
	if dsNum == 0 {
		return ""
	}
	n := rand.Intn(dsNum)
	return DSList[n]
}
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
	//rimuovi il ds dalla lista
	var t []string
	for _, ds := range DSList {
		if ds != dsCrashed {
			t = append(t, ds)
		}
	}
	DSList = t
	fmt.Println("Ho rimosso il ds dalla lista, ora la lista risultante è ")
	fmt.Println(DSList)
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	params := mux.Vars(r)
	ds := chooseDS()
	if ds == "" {
		fmt.Println("Al momento non ci sono ds, riprovare più tardi.")
		json.NewEncoder(w).Encode("Al momento non ci sono ds, riprovare più tardi.")
		return
	}
	fmt.Println("get called: I wanna read " + params["key"] + " on " + ds + "scelto randomicamente") //Acquire url params

	response, err := http.Get("http://" + ds + ":8080/get/" + params["key"]) //Submitting a get request
	if err != nil {
		reportDSCrash(ds)
		fmt.Println("Il ds che ho scelto è crashato, lo rimuovo e lo dico a discovery")
		//fmt.Println(err.Error()) RIMUOVO DA DSLIST IL CRASHATO
		if ds == DSMasterIP {
			DSMasterIP = ""
			fmt.Println("era il master ad essere crashato")
			reportDSMasterCrash()
			return
		}
		for pos, dsToRemove := range DSList {
			if strings.Compare(ds, dsToRemove) == 0 {
				a := DSList[0:pos]
				for _, s := range DSList[pos+1:] { //Rimuovilo
					a = append(a, s)
				}
				DSList = a
			}
		}
		fmt.Println("Il ds è crashato, ora la lista per me è")
		fmt.Println(DSList)
		fmt.Println("Il master non so se c'è ancora, mi risulta essere " + DSMasterIP)
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
	for DSMasterIP == "" {
		fmt.Println("There is not a master. Waiting for a master to come back")
		time.Sleep(3 * time.Second)
		response, _ := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Submitting a put request
		r, _ := ioutil.ReadAll(response.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
	}
	requestJSON, _ := json.Marshal(request)
	if DSMasterIP == "" {
		fmt.Println("Non c'è il master")
		return
	}
	response, err := http.Post("http://"+DSMasterIP+":8080/put", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	if err != nil {
		reportDSMasterCrash()
		//fmt.Println(err.Error())
		for pos, dsToRemove := range DSList {
			if strings.Compare(DSMasterIP, dsToRemove) == 0 {
				a := DSList[0:pos]
				for _, s := range DSList[pos+1:] { //Rimuovilo
					a = append(a, s)
				}
				DSList = a
			}
		}
		DSMasterIP = ""
		response, err := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Submitting a put request
		for err != nil {
			fmt.Println("Discovery crashato, aspetto che torna")
			time.Sleep(3 * time.Second)
			response, err = http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Submitting a put request

		}
		r, _ := ioutil.ReadAll(response.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
		json.NewEncoder(w).Encode("Master crashed. Try again later.")
		return
	}
	responseFromDS, _ := ioutil.ReadAll(response.Body) //Receiving http response
	json.NewEncoder(w).Encode(string(responseFromDS))
}

func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")

	fileToRemove := analyzeRequest(r)
	var request string = fileToRemove //Build the request in a particular format
	fmt.Println("del called: I wanna remove " + fileToRemove + " on " + DSMasterIP)
	for DSMasterIP == "" {
		fmt.Println("There is not a master. Waiting for a master to come back")
		time.Sleep(3 * time.Second)
		response, _ := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Submitting a put request
		r, _ := ioutil.ReadAll(response.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
	}
	requestJSON, _ := json.Marshal(request)
	response, err := http.Post("http://"+DSMasterIP+":8080/del", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	if err != nil {
		fmt.Println(err.Error())
		reportDSMasterCrash()
		for pos, dsToRemove := range DSList {
			if strings.Compare(DSMasterIP, dsToRemove) == 0 {
				a := DSList[0:pos]
				for _, s := range DSList[pos+1:] { //Rimuovilo
					a = append(a, s)
				}
				DSList = a
			}
		}
		DSMasterIP = ""
		response, err := http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Submitting a put request
		for err != nil {
			fmt.Println("Discovery crashato, aspetto che torna")
			time.Sleep(3 * time.Second)
			response, err = http.Post("http://"+DiscoveryIP+":8080/whoisMaster", "application/json", nil) //Submitting a put request

		}
		r, _ := ioutil.ReadAll(response.Body)
		DSMasterIP = string(r)
		DSMasterIP = strings.ReplaceAll(DSMasterIP, "\"", "")
		DSMasterIP = strings.Replace(DSMasterIP, "\n", "", -1)
		//fmt.Println(err.Error())
		json.NewEncoder(w).Encode("Master crashed. Try again later.")
		return
	}
	responseFromDS, _ := ioutil.ReadAll(response.Body) //Receiving http response
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
	router.HandleFunc("/addDs", addDs).Methods("POST")
	router.HandleFunc("/removeDs", removeDs).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router)) //Listen and serve requests on port 8080
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
func addDs(w http.ResponseWriter, r *http.Request) {

	req := analyzeRequest(r)
	if !isInlist(req, DSList) {
		DSList = append(DSList, req)
	}
	fmt.Println("Aggiunta nuova replica: ora l'insieme dei ds è")
	fmt.Println(DSList)
}
func isInlist(e string, l []string) bool {
	for _, elem := range l {
		if strings.Compare(e, elem) == 0 {
			return true
		}
	}
	return false
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
	for strings.Compare(string(responseFromDiscovery)[1:len(string(responseFromDiscovery))-2], "|") == 0 {
		fmt.Println("The master is not here yet. I am gonna wait")
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
		if err != nil {
			fmt.Println("Discovery has crashed. Trying to reconnect ...")
			for err != nil {
				response, err = http.Post("http://"+DiscoveryIP+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
			}
		}
		responseFromDiscovery, _ = ioutil.ReadAll(response.Body) //Receiving http response
	}
	fmt.Println("sto cercando di registrarmi, ho ricevuto come dati " + string(responseFromDiscovery))
	var dslist string = (string(responseFromDiscovery[0 : len(string(responseFromDiscovery))-2]))

	dslist = strings.ReplaceAll(dslist, "\\", "")
	dslist = strings.ReplaceAll(dslist, "n", "") //Cleaning the output
	dslist = strings.ReplaceAll(dslist, "\"", "")
	acquireDSList(dslist) //vengono restituiti tutti i ds con il master alla fine
	DSMasterIP = DSList[len(DSList)-1]
	fmt.Println("registration complete: the master is" + DSMasterIP)
	fmt.Println("registration complete: the dslist is")
	fmt.Println(DSList)
}
func acquireDSList(dslist string) {
	var lastindex = 0 //per via delle doppie virgolette iniziali
	for pos, char := range dslist {
		if char == 124 { //quindi se il carattere letto è |
			if !isInlist(dslist[lastindex:pos], DSList) {
				DSList = append(DSList, dslist[lastindex:pos])
			}
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
