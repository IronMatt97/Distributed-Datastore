package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type Object struct {
	key   string "json:'key'"
	value string "json:'value'"
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	params := mux.Vars(r)                       //Ottengo i parametri nella richiesta url
	data, err := ioutil.ReadFile(params["key"]) //Provo a leggere un file con titolo key letto in richiesta
	fmt.Println("Richiesto file " + params["key"])
	//Se non ci riesco ritorna un oggetto vuoto e l'errore
	if err != nil {
		fmt.Println("An error has occurred reading the file.")
		fmt.Println(err)
		return
	}
	//Se ci riesco encoda un nuovo oggetto con titolo e contenuto
	fmt.Print(string(data))
	json.NewEncoder(w).Encode(string(data))
}

func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	body, err := ioutil.ReadAll(r.Body) //Leggi la richiesta
	if err != nil {
		fmt.Print("errore nella lettura della richiesta")
		fmt.Println(err)
	}
	var receivedRequest string
	err = json.Unmarshal([]byte(body), &receivedRequest)
	if err != nil {
		fmt.Println("Error unmarshaling data from request.")
		return
	}
	var info []string = strings.Split(receivedRequest, "|")
	var fileName string = info[0]
	var fileContent string = info[1]

	//@@TODO----------- Controlla che non ci sia già
	/*data, err := ioutil.ReadFile(fileName) //Provo a leggere un file con titolo key letto in richiesta
	if os.IsExist(err) {
		json.NewEncoder(w).Encode("Error: The file you want to create already exists." )
		return
	}------------------------------*/

	//Scrittura effettiva del file
	err = ioutil.WriteFile(fileName, []byte(fileContent), 0777)
	if err != nil {
		fmt.Println("An error has occurred writing the file. ")
		fmt.Println(err)
	}
	json.NewEncoder(w).Encode("The object was successfully uploaded")
}

func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
	params := mux.Vars(r)
	err := os.Remove(params["key"]) // remove a single file
	if err != nil {
		fmt.Println(err)
	}
	json.NewEncoder(w).Encode("The object was successfully removed.")
}

func main() {
	//Router init - il := serve a fargli capire il tipo da quello che legge dopo cosi da non fare int
	router := mux.NewRouter()
	//Handlers/Endpoints del routes
	router.HandleFunc("/put", put).Methods("POST")
	router.HandleFunc("/get/{key}", get).Methods("GET")
	//router.HandleFunc("/getAll",getAll2).Methods("GET")
	router.HandleFunc("/delete/{key}", del).Methods("DELETE")
	//Gestione della connessione dai nodi, log Fatal serve a dare errore se non va a buon fine
	log.Fatal(http.ListenAndServe(":8000", router))
}
