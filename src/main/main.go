package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"log"
	"regexp"
	"main/data"
	"strconv"
	"fmt"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/query/factorio/{address}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)

		vars := mux.Vars(r)
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "    ")

		compile, _ := regexp.Compile("^([^:\n]+)(?::([0-9]+))?$")
		result := compile.FindAllStringSubmatch(vars["address"], 2)

		address := result[0][1]
		port, _ := strconv.Atoi(result[0][2])

		serverData := data.FactorioServerData{
			Address: address,
			Port:    port,
		}


		query := serverData.QueryServer()

		encoder.Encode(query)
	})

	log.Fatal(http.ListenAndServe(":3080", router))
}
