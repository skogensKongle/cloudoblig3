package main

import (
    "fmt"
    "net/http"
    "os"
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

  //Checking Connection to mongodb
  session, err := mgo.Dial("localhost")
  if err != nil {
    log.Fatal("Could not connect to the mongoDB server")
  }
  Init(session.DB("cloudoblig3", router)
  fmt.Println("listening...")
  //err := http.ListenAndServe(":3000", router)
  session, err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
  if err != nil {
    panic(err)
  }
}


//+++++++++++++++++++++++++ Init function +++++++++++++++++++++++++++++++
func Init(db *mgo.Database, r *mux.Router) {

database = db
  http.Handle("/", router)
  getRates(&mongoRates)
	/*database = db
	// Cron jobs
	c := cron.New()
	getCronData()
	c.AddFunc("@every 24h", getCronData)
	c.Start()

	r.HandleFunc("/", handlerGet).Methods("GET")
	r.HandleFunc("/", handlerPost).Methods("POST")

	r.HandleFunc("/latest", getLatest).Methods("POST")
	r.HandleFunc("/evaluationtrigger", evaluateTrigger).Methods("POST")
	r.HandleFunc("/average", getAverage).Methods("POST")

	r.HandleFunc("/{id}", getWebhook).Methods("GET")
	r.HandleFunc("/{id}", deleteWebhook).Methods("DELETE")*/
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
