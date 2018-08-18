package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/text/encoding/charmap"
)

var hashCurrency map[string]crbValute

type crbValute struct {
	CharCode string
	Nominal  int
	Name     string
	Value    string
}

type crbValcurs struct {
	ValCurs xml.Name
	Valute  []crbValute
}

func crbXMLResponse() (*crbValcurs, error) {
	print("start crbXMLResponse\n")

	xmlFormat := new(crbValcurs)

	con := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 10,
	}

	resp, err := con.Get("http://www.cbr.ru/scripts/XML_daily.asp")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d := xml.NewDecoder(resp.Body)
	d.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	}

	err = d.Decode(xmlFormat)
	if err != nil {
		return nil, err
	}
	print("end crbXMLResponse\n")
	return xmlFormat, nil
}

func reqCurrency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept")
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()
	var slice = []crbValute{}
	for _, val := range r.Form {
		for _, val2 := range val {
			slice = append(slice, hashCurrency[val2])
		}
	}
	d := make(map[string][]crbValute)
	d["currency"] = slice
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

func monitoringCurrency() {
	for {
		structCurrency, err := crbXMLResponse()
		if err != nil {
			log.Fatal(err)
		}

		for _, val := range structCurrency.Valute {
			hashCurrency[val.CharCode] = val
		}
		print("timeout gorutin\n")
		time.Sleep(time.Minute * 5)
	}
}

func main() {
	hashCurrency = make(map[string]crbValute)

	go monitoringCurrency()

	r := mux.NewRouter()

	r.HandleFunc("/api/currency", reqCurrency).Methods("POST")

	print("Start session\n")

	log.Fatal(http.ListenAndServe(":8081", r))
}
