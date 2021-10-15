package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func read() {
	fmt.Print("Inserisci il nome del file che vuoi leggere: ")
	fileToRead, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	fileToRead = fileToRead[0:len(fileToRead)-1] + "" //Ritaglia la stringa togliendo \n
	response, err := http.Get("http://localhost:8000/get/" + fileToRead)
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		return
	}
	responseFromAPI, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed acquiring API response.")
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(responseFromAPI))
}

func write() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Inserisci il nome del file che vuoi scrivere: ")
	fileName, _ := reader.ReadString('\n')
	fileName = fileName[0:len(fileName)-1] + "" //Ritaglia la stringa togliendo \n
	fmt.Print("Inserisci il contenuto del file che vuoi scrivere: ")
	fileContent, _ := reader.ReadString('\n')
	fileContent = fileContent[0:len(fileContent)-1] + "" //Ritaglia la stringa togliendo \n
	if strings.Contains(fileName, "|") || strings.Contains(fileContent, "|") {
		fmt.Println("Il carattere '|' non può essere inserito. ")
		return
	}
	var request string = fileName + "|" + fileContent
	requestJSON, _ := json.Marshal(request)
	responseAPI, err := http.Post("http://localhost:8000/put", "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		return
	}
	fmt.Println(responseAPI)
}

func main() {
	//Per sempre
	for true {
		//Presenta l'applicativo ed ottieni la scelta dell'utente
		t := time.Now()
		h, m, s := t.Clock()
		fmt.Println("Benvenuto, sono le ", h, ":", m, ":", s, ", quale operazione vuoi eseguire?")
		fmt.Println("1) leggere un file")
		fmt.Println("2) scrivere un file")
		fmt.Println("3) eliminare un file")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Inserire il numero dell'azione scelta: ")
		action, _ := reader.ReadString('\n')
		action = action[0:len(action)-1] + "" //Ritaglia la stringa togliendo \n
		switch {
		case action == "1":
			read()
		case action == "2":
			write()
		/*case "3":
		del*/
		default:
			fmt.Println("Scelta non valida.")
		}

		fmt.Println("Premi invio per sottomettere una nuova richiesta, o esegui un'interrupt per uscire.")
		fmt.Scanln() // wait for Enter Key

	}
}
