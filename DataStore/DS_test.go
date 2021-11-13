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
