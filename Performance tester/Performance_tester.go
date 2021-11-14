package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var systemLoad = 100                           //Numero di richieste con cui caricare il sistema
var requestsInterval = 1000 * time.Millisecond //Tempo di attesa tra una richiesta e la successiva
var mockListSize = 100                         //Dimensione dell'array di stringhe casuali per il test
var DiscoveryAddress = "172.17.0.2"            //Indirizzo del Discovery da contattare per effettuare il test
var requestLatencyList []time.Duration         //Lista contenente i tempi di servizio di ogni richiesta
var routines sync.WaitGroup                    //Waitgroup per dare il tempo alle goroutines di chiudersi
var mutex sync.Mutex

func main() {

	fmt.Println("Testing program initialized.")
	testSeed := int64(time.Now().Second())
	rand.Seed(testSeed) //Il seed viene modificato per garantire scenari di testing differenziati
	fmt.Print("Extracted seed: ")
	fmt.Println(testSeed)
	mockStringList := buildMockList() //Per prima cosa viene costruito l'array con numeri casuali

	prepareLocalFiles(mockStringList) //Prepara dei file mock già presenti nel sistema

	for request := 0; request < systemLoad; request++ { //Invio di ogni richiesta con probabilità differenziata
		routines.Add(1)      //Aggiungi la goroutine al waitgroup
		op := rand.Intn(100) //Rispetto a 100 numeri che possono essere estratti, si selezionano determinati intervalli aleatori
		if op < 15 {
			go putOp(mockStringList[rand.Intn(mockListSize)]) //Invio di una put operation
		} else {
			go getOp(mockStringList[rand.Intn(mockListSize)]) //Invio di una get operation
		}
		time.Sleep(requestsInterval) //Attendi prima di sottomettere la successiva richiesta
	}
	routines.Wait() //Questo serve per dare il tempo alle goroutine di chiudersi, cosi da aggiornare la struttura dati
	time.Sleep(3 * time.Second)
	reportTestStats() //Trascrivi le statistiche raccolte all'interno di un apposito file

}

//Funzione per simulare una get operation al sistema
func getOp(item string) {
	defer routines.Done()                                          //Alla fine rimuovi la goroutine dal waitgroup
	initialClock := time.Now()                                     //Acquisisci il tempo iniziale per misurare i tempi nel sistema
	api := requireAPI()                                            //Contatta il discovery per ottenere l'api da utilizzare
	response, _ := http.Get("http://" + api + ":8080/get/" + item) //Invia la richiesta all'API
	ioutil.ReadAll(response.Body)                                  //Ottieni la risposta dall'API
	latency := time.Since(initialClock)
	mutex.Lock()
	requestLatencyList = append(requestLatencyList, latency) //Salva la latenza della richiesta
	mutex.Unlock()
}

//Funzione per simulare una put operation al sistema
func putOp(item string) {
	defer routines.Done()                                                                                 //Alla fine rimuovi la goroutine dal waitgroup
	initialClock := time.Now()                                                                            //Acquisisci il tempo iniziale per misurare i tempi nel sistema
	api := requireAPI()                                                                                   //Contatta il discovery per ottenere l'api da utilizzare
	requestJSON, _ := json.Marshal(item + "|" + item)                                                     //Costruisci la richiesta all'API nel formato atteso
	response, _ := http.Post("http://"+api+":8080/put", "application/json", bytes.NewBuffer(requestJSON)) //Invia la richiesta all'API
	ioutil.ReadAll(response.Body)                                                                         //Ottieni la risposta dall'API
	latency := time.Since(initialClock)
	mutex.Lock()
	requestLatencyList = append(requestLatencyList, latency) //Salva la latenza della richiesta
	mutex.Unlock()
}

//Funzione che contatta il discovery per farsi consegnare una API da usare per ogni richiesta
func requireAPI() string {
	requestJSON, _ := json.Marshal("client")
	response, _ := http.Post("http://"+DiscoveryAddress+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body)
	return cleanString(string(responseFromDiscovery[1 : len(string(responseFromDiscovery))-2]))
}

//Funzione per collocare nel sistema dei file prima di testarlo
func prepareLocalFiles(list []string) {
	api := requireAPI()
	for request := 0; request < systemLoad; request++ { //Lancia qualche richiesta http per scrivere files nel sistema.
		item := list[rand.Intn(mockListSize)]
		requestJSON, _ := json.Marshal(item + "|" + item)
		http.Post("http://"+api+":8080/put", "application/json", bytes.NewBuffer(requestJSON))
	}

}

//Funzione che trascrive le statistiche estratte all'interno di un file CSV per studiarne l'andamento
func reportTestStats() {
	f, _ := os.OpenFile("SystemPerformanceLog.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	f.WriteString("Request;Latency;\n")
	var counter = 1
	for i := 0; i < len(requestLatencyList); i++ {
		str := fmt.Sprint(requestLatencyList[i])
		f.WriteString(fmt.Sprint(counter))
		f.WriteString(";")
		f.WriteString(str[:len(str)-2]) //Rimuovi l'unità di misura
		f.WriteString(";\n")
		counter++
	}
}

//Funzione che costruisce la lista di stringhe mock per il test
func buildMockList() []string {

	var mockList []string
	fmt.Println("Building mock list...")
	for elem := 0; elem < mockListSize; elem++ {
		mockList = append(mockList, fmt.Sprint(rand.Intn(10000)))
	}
	fmt.Println("Mock list generated.")
	return mockList

}

//Funzione dedita alla pulizia delle stringhe ricevute dal sistema
func cleanString(s string) string {
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "n", "")
	s = strings.ReplaceAll(s, "\"", "")
	return s
}
