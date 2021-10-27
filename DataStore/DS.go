package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var Master = false

func main() {
	register()
}

func register() {
	requestJSON, _ := json.Marshal("datastore")
	response, err := http.Post("http://localhost:8000/register", "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the Discovery node.")
		fmt.Println(err.Error())
		return
	}
	responseFromAPI, _ := ioutil.ReadAll(response.Body) //Receiving http response
	fmt.Println(string(responseFromAPI))
	//Controlla se la rispostaè  positiva altrimenti ricomincia il main tipo if ok return else richiama register()
}
