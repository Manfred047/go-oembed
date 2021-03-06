package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"net/http"

	"github.com/badoux/goscraper"
	"github.com/dyatlov/go-oembed/oembed"
	"encoding/json"
	"strconv"
)

func metaHandler(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["u"]

	if !ok || len(keys[0]) < 1 {
		http.Error(w, "Url Param 'u' is missing", http.StatusNotFound)
		return
	}

	data, err := ioutil.ReadFile("./providers.json")

	if err != nil {
		setError("ERROR: No se puede leer providers.json, status: " + err.Error())
		metaSecondHandler(w, r)
	}

	oe := oembed.NewOembed()
	oe.ParseProviders(bytes.NewReader(data))

	url := keys[0]

	url = strings.Trim(url, "\r\n")

	if url == "" {
		http.Error(w, "Not found", http.StatusNotFound)
	}

	item := oe.FindItem(url)

	if item != nil {
		info, err := item.FetchOembed(oembed.Options{URL: url, AcceptLanguage: "es-MX"})
		if err != nil {
			setError("ERROR: No se pueden obtener los datos desde oembed, status: " + err.Error())
			metaSecondHandler(w, r)
		} else {
			if info.Status >= 300 {
				code := strconv.Itoa(info.Status)
				setError("ERROR: No se pueden obtener los datos desde url, status: " + code)
				metaSecondHandler(w, r)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				json.NewEncoder(w).Encode(info)
			}
		}
	} else {
		setError("ERROR: no se encuentra el item en oembed")
		metaSecondHandler(w, r)
	}
}

type Preview struct {
	From string `json:"from"`
	Url   	  string   `json:"url"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Images      []string   `json:"images"`
}

func metaSecondHandler(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()["u"]
	s, err := goscraper.Scrape(keys[0], 5)
	if err != nil {
		setError("ERROR: No se puede generar la vista, status: " + err.Error())
		http.Error(w, "can't generate preview", http.StatusBadRequest)
		return
	}

	var pvw Preview
	pvw.From = "secundary"
	pvw.Url = s.Preview.Link
	pvw.Title = s.Preview.Title
	pvw.Description = s.Preview.Description
	pvw.Images = s.Preview.Images

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(pvw)
}

func main() {
	http.HandleFunc("/", metaHandler)

	http.ListenAndServe(GetPort(), nil)
}

func setError(error string) {
	var dev = os.Getenv("APP_ENV")
	if dev == "local" {
		fmt.Println(error)
	}
}

func GetPort() string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "4747"
		setError("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	fmt.Println("Escuchando el puerto: " + port)
	return ":" + port
}