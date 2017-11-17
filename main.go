package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"


    "github.com/gorilla/mux"
  	"gopkg.in/mgo.v2"
)

// struct for saving Database
type Mongo struct {
	DatabaseURL     string
	DatabaseName    string
	MongoCollection string
}
//The database i use
var mongoRates = Mongo{DatabaseURL: "mongodb://stisoe:1234@ds113136.mlab.com:13136/cloudoblig3", DatabaseName: "cloudoblig3", MongoCollection: "rates"}
var mongoTickets = Mongo{DatabaseURL: "mongodb://stisoe:1234@ds113136.mlab.com:13136/cloudoblig3", DatabaseName: "cloudoblig3", MongoCollection: "tickets"}

//Getting stuff from fixer.io
type FromFixer struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float32 `json:"rates"`
}

var database *mgo.Database

func main() {

	router := mux.NewRouter()

  http.Handle("/", router)
  getRates(&mongoRates)

  fmt.Println("listening...")
  //err := http.ListenAndServe(":3000", router)
  err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
  if err != nil {
    panic(err)
  }
}


//++++++++++++++++ fetching rates from fixer ++++++++++++++++++++++++++++

func getRates(db *Mongo) {

	var getAllRates FromFixer

	fmt.Print(db.DatabaseURL)
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	//Fetch response form url
	var url = "https://api.fixer.io/latest"

	repo, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer repo.Body.Close()

	//Grab body from Response

	body, err := ioutil.ReadAll(repo.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &getAllRates)
	if err != nil {
		panic(err)
	}

	err = session.DB(db.DatabaseName).C(db.MongoCollection).Insert(getAllRates)
	if err != nil {
		fmt.Printf("Error in Insert(): %v", err.Error())
	}

  //testing if i get rates
  fmt.Print(getAllRates)
}
