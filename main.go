// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	config.IsHTTP = true
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

var config = GetConfig()

func main() {
	InitDB()
	entryPoint := fmt.Sprintf("/%s", config.EntryPoint)
	http.HandleFunc(entryPoint, ProcessSolution)
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
