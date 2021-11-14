package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

//Funzione di inizializzazione: il client contatta il discovery per conoscere l'API con la quale dovrà comunicare
func TestRegister(t *testing.T) {
	register()
}

//Funzione necessaria per comunicare al discovery il crash dell'API in uso
func TestApicrash(t *testing.T) {
	apicrash()
}

//Funzione di controllo della validità delle stringhe inserite
func TestIsStringIllegal(t *testing.T) {
	isStringIllegal("ciao")
	isStringIllegal("ciao.")
	isStringIllegal("ciao\\")
}

//Funzione di stampa a console di presentazione dell'applicativo
func TestClientInit(t *testing.T) {
	clientInit()
}

//Funzione che implementa l'attesa del client per le azioni successive
func TestWaitForNextAction(t *testing.T) {
	waitForNextAction()
}

//Funzione dedita alla pulizia delle stringhe ricevute dal sistema
func TestCleanResponse(t *testing.T) {
	mock, _ := json.Marshal("ciao")
	cleanResponse(mock)
}

//Funzione dedita alla pulizia delle stringhe ricevute dal sistema
func TestCleanString(t *testing.T) {
	a := cleanString("\\Stringa\"DiProva")
	fmt.Println(a)
}
