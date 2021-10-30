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
	"strings"
	"time"
)

var APIaddress string = "localhost"

//TODO IMPLEMENTA APIcrash
func read() {
	fmt.Print("Name the file that you would like to read: ")
	fileToRead := acquireString()
	response, err := http.Get("http://" + APIaddress + ":8000/get/" + fileToRead) //Submitting a get request
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		return
	}
	responseFromAPI, err := ioutil.ReadAll(response.Body) //Receiving http response
	if err != nil {
		fmt.Println("An error has occurred trying to read the requested file.")
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(responseFromAPI))
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
	response, err := http.Post("http://"+APIaddress+":8000/put", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a put request
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		return
	}
	responseFromAPI, err := ioutil.ReadAll(response.Body) //Receiving http response
	if err != nil {
		fmt.Println("An error has occurred trying to read the requested file.")
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(responseFromAPI))
}

func del() {
	fmt.Print("Insert the name of the file that you would like to delete: ")
	fileToRemove := acquireString()
	if isStringIllegal(fileToRemove) {
		return
	}
	requestJSON, _ := json.Marshal(fileToRemove)
	response, err := http.Post("http://"+APIaddress+":8000/del", "application/json", bytes.NewBuffer(requestJSON)) //Submitting a delete request
	if err != nil {
		fmt.Println("An error has occurred trying to estabilish a connection with the API.")
		fmt.Println(err.Error())
		return
	}
	responseFromAPI, err := ioutil.ReadAll(response.Body) //Receiving http response
	if err != nil {
		fmt.Println("An error has occurred trying to remove the requested file.")
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(responseFromAPI))
}

func main() {

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

func acquireString() string {
	stdin, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	stdin = stdin[0:len(stdin)-1] + "" //Ritaglia la stringa togliendo \n
	return stdin
}
func isStringIllegal(s string) bool {
	if strings.Contains(s, "|") || strings.Contains(s, ".") || strings.Contains(s, "/") || strings.Compare(s, "") == 0 {
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
