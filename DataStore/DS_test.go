package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

//Use case: il client vuole leggere un file
func TestGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/get/", nil)
	w := httptest.NewRecorder()
	get(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Use case: il client vuole scrivere un file
func TestPut(t *testing.T) {
	mock, _ := json.Marshal("prova|prova")
	req := httptest.NewRequest(http.MethodPost, "/put", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	put(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Use case: il client vuole eliminare un file
func TestDel(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/del", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	del(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione di registrazione al discovery
func TestRegister(t *testing.T) {
	register()
}

//Funzione utile all'implementazione del parallelismo di goroutines per la put
func TestContactReplicas(t *testing.T) {
	mockDelete, _ := json.Marshal("prova")
	mockPut, _ := json.Marshal("prova|prova")
	contactReplicas("172.17.0.3", mockDelete, "delete")
	contactReplicas("172.17.0.3", mockPut, "put")
}

//Funzione di utility per la rimozione di un Datastore dalla lista
func TestRemoveDSFromList(t *testing.T) {
	removeDSFromList("prova")
}

//Funzione che contatta il Discovery per avvisarlo del crash di un Datastore replica
func TestReportDSCrash(t *testing.T) {
	reportDSCrash("prova")
}

//Funzione per la rimozione di un datastore dalla lista, chiamabile dal Discovery
func TestRemoveDs(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/removeDs", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	removeDs(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione per aggiungere un Datastore alla lista, chiamabile dal Discovery
func TestAddDs(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/addDs", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	addDs(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione chiamabile dal Discovery per l'elezione del Master
func TestBecomeMaster(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/becomeMaster", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	becomeMaster(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione di utility per il controllo di appartenenza alla lista di una stringa
func TestIsInlist(t *testing.T) {
	var mocklist []string
	mocklist = append(mocklist, "prova")
	mocklist = append(mocklist, "prova2")
	isInlist("prova", mocklist)
}

//Funzione di utility per la pulizia delle stringhe
func TestCleanResponse(t *testing.T) {
	mock, _ := json.Marshal("prova")
	cleanResponse(mock)
}

//Funzione usata dalle repliche appena unite al sistema, che resettano lo stato dei file precedente prima di riacquisirlo dal Master
func TestFlushLocalfiles(t *testing.T) {
	flushLocalfiles()
}

//Funzione di reallineamento della replica appena unita al sistema
func TestGetDataUntilNow(t *testing.T) {
	getDataUntilNow()
}

//Funzione che le repliche chiamano sul Master per farsi consegnare la lista dei file già presente per reallinearsi
func TestAlignNewReplica(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/getData", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	alignNewReplica(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione di utility per la costruzione della risposta del DSMaster ai DS replica quando deve consegnargli la lista di file già presenti
func TestPrepareDataList(t *testing.T) {
	prepareDataList()
}

//Funzione di utility che crea la lista dei Datastore a partire dalla risposta nel formato differente di un altro nodo
func TestAcquireDSList(t *testing.T) {
	acquireDSList("prova1|prova2|prova3")
}

//Funzione di utility per convertire la richiesta in un formato leggibile
func TestAnalyzeRequest(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodGet, "/get/", bytes.NewBuffer(mock))
	analyzeRequest(req)
}
