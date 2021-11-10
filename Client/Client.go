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

var APIaddress string = "" //Cambia questo con il balancer delle api
var DiscoveryAddress string = "172.17.0.2"

func read() {
	fmt.Print("Name the file that you would like to read: ")
	fileToRead := acquireString()
	response, err := http.Get("http://" + APIaddress + ":8080/get/" + fileToRead) //Submitting a get request
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		fmt.Println("Sto comunicando che la api è crashata")
		apicrash()
		return
	}
	responseFromAPI, err := ioutil.ReadAll(response.Body) //Receiving http response
	if err != nil {
		fmt.Println("An error has occurred trying to read the requested file.")
		fmt.Println(err.Error())
		return
	}
	fmt.Println("L'api mi ha risposto " + string(responseFromAPI))
	resp := cleanResponse(responseFromAPI)
	fmt.Println("Ho fatto il clean in " + resp)
	fmt.Println(resp)
}

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
	var request string = fileName + "|" + fileContent                                                              //Build the request in a particular format
	requestJSON, _ := json.Marshal(request)                                                                        //Marshal the request
	response, err := http.Post("http://"+APIaddress+":8080/put", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		fmt.Println("Sto comunicando che la api è crashata")
		apicrash()
		return
	}
	responseFromAPI, err := ioutil.ReadAll(response.Body) //Receiving http response
	if err != nil {
		fmt.Println("An error has occurred trying to read the requested file.")
		fmt.Println(err.Error())
		return
	}
	resp := cleanResponse(responseFromAPI)
	fmt.Println("L api mi ha risposto: ")
	fmt.Println(resp)
}

func del() {
	fmt.Print("Insert the name of the file that you would like to delete: ")
	fileToRemove := acquireString()
	if isStringIllegal(fileToRemove) {
		return
	}
	requestJSON, _ := json.Marshal(fileToRemove)
	response, err := http.Post("http://"+APIaddress+":8080/del", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a delete request
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		fmt.Println("Sto comunicando che la api è crashata")
		apicrash()
		return
	}
	responseFromAPI, err := ioutil.ReadAll(response.Body) //Receiving http response
	if err != nil {
		fmt.Println("An error has occurred trying to remove the requested file.")
		fmt.Println(err.Error())
		return
	}
	resp := cleanResponse(responseFromAPI)
	fmt.Println("Ho capito come risposta dall'api")
	fmt.Println(resp)
}

func main() {

	register()
	for {
		clientInit()              //Initialize application
		action := acquireString() //Acquire user target action
		switch {
		case action == "1":
			read()
		case action == "2":
			write()
		case action == "3":
			del()
		default:
			fmt.Println("Invalid input. Restarting the program ...")
		}
		waitForNextAction() //Wait for Enter key ...
	}
}

func register() {
	fmt.Println("Sto cercando di registrarmi al discovery")
	requestJSON, _ := json.Marshal("client")
	response, err := http.Post("http://"+DiscoveryAddress+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	for err != nil { //Se fallisce riprova ogni 3 secondi
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node.")
		fmt.Println(err.Error())
		time.Sleep(3 * time.Second)
		response, err = http.Post("http://"+DiscoveryAddress+":8080/register", "application/json", bytes.NewBuffer(requestJSON))
	}
	responseFromDiscovery, _ := ioutil.ReadAll(response.Body) //Receiving http response
	fmt.Println("The discovery answered: " + string(responseFromDiscovery) + " devo stare attento non sia noapi")
	if string(responseFromDiscovery) == "noapi" {
		fmt.Println("there are no api to communicate with. retry later.")
		os.Exit(-1)
	}

	APIaddress = (string(responseFromDiscovery[1 : len(string(responseFromDiscovery))-2]))
	APIaddress = strings.ReplaceAll(APIaddress, "\\", "")
	APIaddress = strings.ReplaceAll(APIaddress, "n", "") //Cleaning the output
	APIaddress = strings.ReplaceAll(APIaddress, "\"", "")
	fmt.Println("Ora acquisiro della stringa suddetta " + APIaddress)

	fmt.Println("registration complete: the api is" + APIaddress)
}
func apicrash() {
	fmt.Println("The api has crashed. I am gonna ask discovery a new api to use")
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
		fmt.Println("There are no rest api available. retry later")
		os.Exit(-1)
	}
	APIaddress = string(responseFromDiscovery)[1 : len(responseFromDiscovery)-2]
	fmt.Println("The discovery answered the new api: " + APIaddress)
}

func acquireString() string {
	stdin, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	stdin = stdin[0:len(stdin)-1] + "" //Ritaglia la stringa togliendo \n
	return stdin
}
func isStringIllegal(s string) bool {
	if strings.Contains(s, "|") || strings.Contains(s, ".") || strings.Contains(s, "/") || strings.Compare(s, "") == 0 || strings.Compare(s, "DS") == 0 {
		fmt.Println("The inserted input is not admitted. Avoid using '.','|' or '/'. Restarting the program ...")
		return true
	}
	return false
}
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
func waitForNextAction() {
	fmt.Println("Press Enter to continue, or execute an interrupt to leave.")
	bufio.NewReader(os.Stdin).ReadString('\n')
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
