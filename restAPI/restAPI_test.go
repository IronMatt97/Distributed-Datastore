package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

//Use case: il client vuole effettuare una get
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

//Funzione di utility che viene chiamata dal Discovery per informare questo nodo API dell'elezione di un nuovo DS Master
func TestChangeDSMasterOnCrash(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/changeMaster", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	changeDSMasterOnCrash(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione di recovery chiamata dal Datastore quando riparte per riaggiornarsi sullo stato del sistema
func TestWhoIsMaster(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/whoIMaster", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	changeDSMasterOnCrash(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione chiamata dal Discovery per aggiungere un nuovo Datastore alla lista
func TestAddDs(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/addDs", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	addDs(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione di utility per rimuovere un Datastore dalla lista
func TestRemoveDs(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodPost, "/removeDs", bytes.NewBuffer(mock))
	w := httptest.NewRecorder()
	removeDs(w, req)
	res := w.Result()
	fmt.Println(res)
}

//Funzione per la scelta di un DS tra quelli presenti al fine di fare la get
func TestChooseDS(t *testing.T) {
	chooseDS()
}

//Funzione che comunica al nodo discovery il crash del Datastore Master
func TestReportDSMasterCrash(t *testing.T) {
	reportDSMasterCrash()
}

//Funzione che comunica al nodo discovery il crash di un Datastore
func TestReportDSCrash(t *testing.T) {
	reportDSCrash("prova")
}

//Funzione di utility che rimuove un Datastore dalla lista
func TestRemoveDSFromList(t *testing.T) {
	removeDSFromList("prova")
}

//Funzione di utility per vedere se una stringa appartiene alla lista
func TestIsInlist(t *testing.T) {
	var mocklist []string
	mocklist = append(mocklist, "prova")
	mocklist = append(mocklist, "prova2")
	isInlist("prova", mocklist)
}

//Funzione di utility per estrapolare la lista dei DS dalla risposta del discovery
func TestAcquireDSList(t *testing.T) {
	acquireDSList("prova1|prova2")
}

//Funzione di utility per l'estrapolazione della stringa dalla sequenza di bytes ricevuti
func TestAnalyzeRequest(t *testing.T) {
	mock, _ := json.Marshal("prova")
	req := httptest.NewRequest(http.MethodGet, "/get/", bytes.NewBuffer(mock))
	analyzeRequest(req)
}
