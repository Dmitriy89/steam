package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gorilla/mux"
)

var hashGame map[int]apps

type apps struct {
	Appid int    `json:"appid"`
	Name  string `json:"name"`
}

type listgame struct {
	Applist struct {
		Apps []apps `json:"apps"`
	} `json:"applist"`
}

func limitGame(b []byte) []byte {
	limit := make([]byte, 988)
	copy(limit, b)
	limit = append(limit, b[len(b)-3:]...)
	return limit
}

func getInfoGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept")
	w.Header().Set("Content-Type", "application/json")
	p, _ := infoGame(r.FormValue("idgame"))

	var slice = make([]int, 0)
	slice = append(slice, p)
	d := make(map[string][]int)
	d["infogame"] = slice
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

func infoGame(n string) (int, error) {

	con := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 10,
	}

	resp, err := con.Get("https://store.steampowered.com/api/appdetails?appids=" + n)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	b1, err := jsonparser.GetInt(body, n, "data", "price_overview", "final")
	if err != nil {
		return 0, err
	}

	priceStr := strconv.Itoa(int(b1))
	priceInt, err := strconv.Atoi(priceStr[:len(priceStr)-2])
	if err != nil {
		return 0, err
	}
	return priceInt, err
}

func listGame() (*listgame, error) {

	list := new(listgame)

	con := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 10,
	}

	resp, err := con.Get("http://api.steampowered.com/ISteamApps/GetAppList/v2")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	limitList := limitGame(body)

	err = json.Unmarshal(limitList, list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func getListGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept")
	w.Header().Set("Content-Type", "application/json")
	var slice = []apps{}
	for key := range hashGame {
		slice = append(slice, hashGame[key])
	}
	d := make(map[string][]apps)
	d["game"] = slice
	err := json.NewEncoder(w).Encode(d)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

func main() {

	hashGame = make(map[int]apps)

	list, err := listGame()
	if err != nil {
		log.Fatal(err)
	}

	for _, val := range list.Applist.Apps {
		hashGame[val.Appid] = val
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/listgame", getListGame).Methods("POST")
	r.HandleFunc("/api/infogame", getInfoGame).Methods("POST")

	print("Start session\n")

	log.Fatal(http.ListenAndServe(":8083", r))
}
