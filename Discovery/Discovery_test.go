package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	//Libreria aggiuntiva presa da github che permette di utilizzare facilmente un servizio di listen and serve su una porta
)

//Funzione per scegliere una API da assegnare al client quando si unisce al sistema
func TestChooseAPI(t *testing.T) {
	chooseAPI()
}

//Funzione per la rimozione di un datastore dalla lista in seguito ad un crash ad esempio
func TestRemoveDSFromList(t *testing.T) {
	removeDSFromList("prova")
}

//Funzione per la rimozione di un datastore per via di un crash, chiamabile dagli altri nodi
func TestDsCrash(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/dsCrash", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	dsCrash(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione dedita all'elezione di un nuovo master in seguito al crash dell'attuale, chiamabile dall'esterno
func TestDsMasterCrash(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/dsMasterCrash", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	dsMasterCrash(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione contenente la logica di elezione del nuovo Master
func TestElectNewMaster(t *testing.T) {
	electNewMaster()
}

//Funzione fondamentale del discovery che serve a registrare qualsiasi altro nodo voglia connettersi al sistema
func TestRegisterNewNode(t *testing.T) {
	mock, _ := json.Marshal("client")
	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(mock))
	dsMasterCrash(w, req)
	fmt.Println(w.Result())
	mock, _ = json.Marshal("restAPI")
	req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(mock))
	dsMasterCrash(w, req)
	fmt.Println(w.Result())
	mock, _ = json.Marshal("datastore")
	req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(mock))
	dsMasterCrash(w, req)
	fmt.Println(w.Result())
}

//Funzione di utility per la rimozione di stringhe da liste
func TesRemoveAPIFromList(t *testing.T) {
	removeAPIFromList("prova")
}

//Funzione chiamabile dal client per informare il Discovery del crash di una API
func TestApicrash(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/apicrash", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	apicrash(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione di utility chiamabile dall'esterno per consegnare l'indirizzo del master su richiesta
func TestWhoIsMaster(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/whoisMaster", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	whoIsMaster(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione di utility per la costruzione della richiesta a partire dal formato lista
func TestBuildDSList(t *testing.T) {
	buildDSList()
}

//Funzione di recovery del Discovery; quando si avvia controlla sempre che non ci fosse già qualcuno nel sistema, per recuperare lo stato
func TestCheckForPrevState(t *testing.T) {
	checkForPrevState()
}

//Funzione di utility per la pulizia di stringhe
func TestCleanResponse(t *testing.T) {
	mock, _ := json.Marshal("prova")
	cleanResponse(mock)
}

//Funzione di utility per scoprire se l'apiList è vuota
func TestApiListEmpty(t *testing.T) {
	apiListEmpty()
}

//Funzione di utility per scoprire se la dsList è vuota
func TestDsListEmpty(t *testing.T) {
	dsListEmpty()
}

//Funzione di utility per controllare se una stringa appartiene alla lista
func TestIsInlist(t *testing.T) {
	var mocklist []string
	mocklist = append(mocklist, "prova")
	mocklist = append(mocklist, "prova2")
	isInlist("prova", mocklist)
}

//Funzione di utility per covertire la sequenza di byte di risposta in una stringa leggibile
func TestAnalyzeRequest(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodGet, "/get/", bytes.NewBuffer(mock))
	analyzeRequest(req)
}

//Funzione di utility per l'acquisizione degli indirizzi ip in liste, sia per i ds che per le api
func TestAcquireIP(t *testing.T) {
	acquireIP("IndirizzoIpDiProva", "restAPI")
	acquireIP("IndirizzoIpDiProva", "datastore")
}
