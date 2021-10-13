package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Object struct {
	key   string "json:'key'"
	value string "json:'value'"
}

type ObjectResponse struct {
	key   string `json:"key"`
	value int    `json:"value"`
}

func read() {
	fmt.Print("Inserisci il nome del file che vuoi leggere: ")
	reader := bufio.NewReader(os.Stdin)
	fileToRead, _ := reader.ReadString('\n')
	fileToRead = fileToRead[0:len(fileToRead)-1] + "" //Ritaglia la stringa togliendo \n
	response, err := http.Get("http://localhost:8000/get/" + fileToRead)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(responseData))
}
func write() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Inserisci il nome del file che vuoi scrivere: ")
	fileName, _ := reader.ReadString('\n')
	fileName = fileName[0:len(fileName)-1] + "" //Ritaglia la stringa togliendo \n
	fmt.Print("Inserisci il contenuto del file che vuoi scrivere: ")
	fileContent, _ := reader.ReadString('\n')
	fileContent = fileContent[0:len(fileContent)-1] + "" //Ritaglia la stringa togliendo \n

	//Necessary information marshaling to pass
	/*var file Object
	file.key = fileName
	file.value = fileContent
	fileJSON, err := json.Marshal(file)
	fmt.Print("File JSON da inviare: ")
	fmt.Println(fileJSON)
	var o Object
	err = json.Unmarshal([]byte(fileJSON), &o)
	*/
	emp_obj := Object{fileName, fileContent}
	emp, _ := json.Marshal(emp_obj)
	fmt.Println(string(emp))

	rbytes := []byte(emp)
	var res ObjectResponse
	json.Unmarshal(rbytes, &res)

	fmt.Println("DECRIPTATO: " + res.key)

	//response, err := http.Post("http://localhost:8000/put", "application/json", bytes.NewBuffer(fileJSON))
	/*if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Client esploso")
		return
	}
	fmt.Println(response)*/
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
