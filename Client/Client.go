package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

var APIaddress string = ""                 //L'indirizzo della API con la quale il client comunicherà
var DiscoveryAddress string = "172.17.0.2" //L'inirizzo del Discovery

func main() {
	register()
	for {
		clientInit()              //Inizializza l'output
		action := acquireString() //Acquisisci l'intento dell'utente
		switch {                  //Esegui l'azione richiesta
		case action == "1":
			read()
		case action == "2":
			write()
		case action == "3":
			del()
		default:
			fmt.Println("Invalid input received. Restarting the program ...")
		}
		waitForNextAction() //Attendi che l'utente prema invio ...
	}
}

//Use case: il client vuole leggere un file
func read() {
	fmt.Print("Name the file that you would like to read: ")
	fileToRead := acquireString()                                                 //Acquisisci il titolo del file dall'utente
	response, err := http.Get("http://" + APIaddress + ":8080/get/" + fileToRead) //Invia la richiesta all'API
	for err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		apicrash() //Informa il Discovery del crash dell'API
		time.Sleep(3 * time.Second)
		response, err = http.Get("http://" + APIaddress + ":8080/get/" + fileToRead)
	}
	responseFromAPI, _ := ioutil.ReadAll(response.Body) //Ottieni la risposta dall'API
	fmt.Println(cleanResponse(responseFromAPI))         //Scrivi in output il contenuto del file richiesto
}

//Use case: il client vuole scrivere un file
func write() {
	fmt.Print("Insert the name of the file that you would like to create: ")
	fileName := acquireString()
	if isStringIllegal(fileName) {
		return
	}
	fmt.Print("Now write the content of the file that you would like to create: ")
	fileContent := acquireString()
	if isStringIllegal(fileContent) {
		return
	}
	requestJSON, _ := json.Marshal(fileName + "|" + fileContent) //Costruisci la richiesta all'API nel formato atteso
	response, err := http.Post("http://"+APIaddress+":8080/put", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		apicrash()
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+APIaddress+":8080/put", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromAPI, _ := ioutil.ReadAll(response.Body)
	fmt.Println(cleanResponse(responseFromAPI))
}

//Use case: il client vuole eliminare un file
func del() {
	fmt.Print("Insert the name of the file that you would like to delete: ")
	fileToRemove := acquireString()
	if isStringIllegal(fileToRemove) {
		return
	}
	requestJSON, _ := json.Marshal(fileToRemove)
	response, err := http.Post("http://"+APIaddress+":8080/del", "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		apicrash()
		response, err = http.Post("http://"+APIaddress+":8080/del", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromAPI, _ := ioutil.ReadAll(response.Body)
	fmt.Println(cleanResponse(responseFromAPI))
}

//Funzione di inizializzazione: il client contatta il discovery per conoscere l'API con la quale dovrà comunicare
func register() {
	fmt.Println("Client node initialized. Trying to contact the discovery service ...")
	time.Sleep(500 * time.Millisecond)       //Latenza simulata
	requestJSON, _ := json.Marshal("client") //Invia la richiesta di join al discovery
	response, err := http.Post("http://"+DiscoveryAddress+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Riprova ogni 3 secondi in caso di errore
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node. Retrying ...")
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryAddress+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body)
	if string(responseFromDiscovery) == "noapi" { //Qualora non ci siano API da contattare, il client si disconnette.
		fmt.Println("There are no API to communicate with at the moment. Sorry for the inconvenience.")
		os.Exit(-1)
	}
	APIaddress = cleanString(string(responseFromDiscovery[1 : len(string(responseFromDiscovery))-2]))
	fmt.Println("The registration process was completed correctly: connection with API " + APIaddress + "estabilished.")
}

//Funzione necessaria per comunicare al discovery il crash dell'API in uso
func apicrash() {
	fmt.Println("Connecting to the discovery node to acquire a new API ...")
	time.Sleep(500 * time.Millisecond)
	requestJSON, _ := json.Marshal(APIaddress)
	response, err := http.Post("http://"+DiscoveryAddress+":8080/apicrash", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Se fallisce riprova ogni 3 secondi
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node. Retrying...")
		fmt.Println(err.Error())
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryAddress+":8080/apicrash", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body) //Receiving http response
	if strings.Compare(string(responseFromDiscovery), "noapi") == 0 {
		fmt.Println("There are no API to communicate with at the moment. Sorry for the inconvenience.")
		os.Exit(-1)
	}
	APIaddress = string(responseFromDiscovery)[1 : len(responseFromDiscovery)-2]
	fmt.Println("Connection with the API " + APIaddress + " estabilished.")
}

//Funzione di acquisizione pulita da standard input
func acquireString() string { //Acquisisci la stringa dall'utente
	stdin, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	stdin = stdin[0:len(stdin)-1] + "" //Ritaglia la stringa togliendo \n
	return stdin
}

//Funzione di controllo della validità delle stringhe inserite
func isStringIllegal(s string) bool {
	if strings.Contains(s, "|") || strings.Contains(s, ".") || strings.Contains(s, "/") || strings.Contains(s, "\\") || strings.Contains(s, "\"") || strings.Compare(s, "") == 0 || strings.Compare(s, "DS") == 0 {
		fmt.Println("The inserted input is not admitted. Avoid using '.','|' or '/'. Restarting the program ...")
		return true
	}
	return false
}

//Funzione di stampa a console di presentazione dell'applicativo
func clientInit() {
	user, _ := user.Current()
	t := time.Now()
	h, m, s := t.Clock()
	fmt.Println("-----------------------------------------------------------------------------------------------")
	fmt.Println("Welcome ", user.Username, ", it's ", h, ":", m, ":", s, ", which action do you want to perform?")
	fmt.Println("1) Read a file")
	fmt.Println("2) Write a file")
	fmt.Println("3) Delete a file")
	fmt.Print("Insert the number of your choice: ")
}

//Funzione che implementa l'attesa del client per le azioni successive
func waitForNextAction() {
	fmt.Println("Press Enter to continue, or execute an interrupt to leave.")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

//Funzione dedita alla pulizia delle stringhe ricevute dal sistema
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

//Funzione dedita alla pulizia delle stringhe ricevute dal sistema
func cleanString(s string) string {
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "n", "")
	s = strings.ReplaceAll(s, "\"", "")
	return s
}
