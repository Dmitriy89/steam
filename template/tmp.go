package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var hashIndex map[int]string

type games struct {
	Game []struct {
		Appid int    `json:"appid"`
		Name  string `json:"name"`
	} `json:"game"`
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "origin, content-type, accept")
	t, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	tmpErr := t.ExecuteTemplate(w, "index", hashIndex)
	if tmpErr != nil {
		fmt.Fprintf(w, tmpErr.Error())
	}

	err = formIDGame(r.FormValue("idgame"))
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
}

func indexGame() error {

	list := new(games)

	con := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 10,
	}

	resp, err := con.Post("http://127.0.0.1:8083/api/listgame", "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, list)
	if err != nil {
		return err
	}

	hashIndex = make(map[int]string)

	for _, val := range list.Game {
		hashIndex[val.Appid] = val.Name
	}

	return nil
}

func formIDGame(n string) error {
	con := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 10,
	}

	resp, err := con.Get("http://127.0.0.1:8083/api/infogame" + n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(body)
	return nil
}

func main() {

	err := indexGame()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", indexPage)

	print("Start session\n")

	log.Fatal(http.ListenAndServe(":8082", r))
}
