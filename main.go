// main.go
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func GetConfig() Config {
	env := "dev"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}
	file_name := fmt.Sprintf("./%s_config.json", env)
	file, err := os.Open(file_name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

var config = GetConfig()

func ProcessCyberCat(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	var err error

	switch r.Method {
	case "GET":
		resp["result"] = "Received GET request"
	case "POST":
		resp["result"] = "Received POST request"
	default:
		err = errors.New("Only GET and POST methods are supported")
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Failed user request, error: %s", err.Error())
		resp["error"] = fmt.Sprintf("Error: %s", err.Error())
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Can't prepare response JSON, error: %s", err.Error())
		jsonResp = []byte(`{"error": "Error happened in response JSON creation"`)
	}
	log.Printf("[RESP]: %s", jsonResp)
	w.Write(jsonResp)
}

func main() {
	InitDB()
	entryPoint := fmt.Sprintf("/%s", config.EntryPoint)
	http.HandleFunc(entryPoint, ProcessSolution)
	http.HandleFunc("/cyber-cat", ProcessCyberCat)
	var err error
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	log.Printf("Started listening on %s%s with HTTP status %t", addr, entryPoint, config.IsHTTP)
	if config.IsHTTP {
		err = http.ListenAndServe(addr, nil)
	} else {
		err = http.ListenAndServeTLS(addr, config.CertFilename, config.KeyFilename, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
