package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	cron "gopkg.in/robfig/cron.v2"
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

//Webhook for incoming Post Requests
type WebHook struct {
	ID              bson.ObjectId `bson:"_id,omitempty"`
	Webhookurl      string        `json:"webhookURL"`
	Basecurrency    string        `json:"baseCurrency"`
	Targetcurrency  string        `json:"targetCurrency"`
	Mintriggervalue float32       `json:"minTriggerValue"`
	Maxtriggervalue float32       `json:"maxTriggerValue"`
}

// LatestRates struct
type LatestRates struct {
	BaseCurrency   string `json:"baseCurrency"`
	TargetCurrency string `json:"targetCurrency"`
}

// Convertion Holds a single from to currency value
type Convertion struct {
	From      string  `json:"from"`
	FromValue float32 `json:"from_value"`
	To        string  `json:"to"`
	ToValue   float32 `json:"to_value"`
	Rate      float32 `json:"rate"`
}

// struct to get json from dialogFlow

type FromDialog struct {
	Result struct {
		Parameters struct {
			BaseCurrency   string `json:"baseCurrency"`
			TargetCurrency string `json:"targetCurrency"`
		} `json:"parameters"`
	} `json:"result"`
}

func main() {

	router := mux.NewRouter()

	daily(&mongoRates)

	http.Handle("/", router)
	router.HandleFunc("/", handlerpost).Methods("POST")
	router.HandleFunc("/average", handlerAver).Methods("POST")
	router.HandleFunc("/latest", handlerlate).Methods("POST")

	router.HandleFunc("/{ID}", handlerEx).Methods("GET")

	router.HandleFunc("/{ID}", handlerDel).Methods("DELETE")

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

//+++++++++++++++++++++++ geting rates ons a day +++++++++++++++++++++++++++++

func daily(db *Mongo) {
	cron := cron.New()
	cron.AddFunc("@daily", func() {
		getRates(&mongoRates)
		fmt.Print("Doing daylies...")
	})
	cron.Start()
}

//++++++++++++++++++++++  add function ++++++++++++++++++++++++++++++++++

func (db *Mongo) add(new WebHook) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	//Handler to DB
	err = session.DB(db.DatabaseName).C(db.MongoCollection).Insert(new)
	if err != nil {
		fmt.Printf("Error in Insert(): %v", err.Error())
	}
}

//+++++++++++++++++++++++++ get function ++++++++++++++++++++++++++++++++
func (db *Mongo) get(keyID string) WebHook {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	id := bson.ObjectIdHex(keyID)
	webhook := WebHook{}
	err = session.DB(db.DatabaseName).C(db.MongoCollection).FindId(id).One(&webhook)
	if err != nil {
		return webhook
	}
	return webhook
}

//+++++++++++++++++++++++++ delete function ++++++++++++++++++++++++++++++++
func (db *Mongo) delete(keyID string) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	id := bson.ObjectIdHex(keyID)

	session.DB(db.DatabaseName).C(db.MongoCollection).RemoveId(id)
}

//+++++++++++++++++++++++++ average +++++++++++++++++++++++++++++++++++++++
func aver(web *LatestRates) float32 {
	session, err := mgo.Dial(mongoRates.DatabaseURL)
	if err != nil {
		panic(err)
	}
	var rates []FromFixer
	err = session.DB(mongoRates.DatabaseName).C(mongoRates.MongoCollection).Find(nil).Sort("-_id").Limit(7).All(&rates)
	if err != nil {
		panic(err)
	}
	var days float32 = 0
	for _, rate := range rates {
		days += rate.Rates[web.TargetCurrency]
	}
	return (days / float32(len(rates)))
}

//++++++++++++++++++++++++++++ latest ++++++++++++++++++++++++++++++++++++++++

func latest(l *LatestRates) Convertion {
	session, err := mgo.Dial(mongoRates.DatabaseURL)
	if err != nil {
		panic(err)
	}
	var rates FromFixer
	err = session.DB(mongoRates.DatabaseName).C(mongoRates.MongoCollection).Find(nil).Sort("-_id").One(&rates)
	if err != nil {
		panic(err)
	}
	rate := rates.As(l.BaseCurrency).To(l.TargetCurrency)
	return rate
}

//++++++++++++++++++++++++++++ AS ++++++++++++++++++++++++++++++++++++++++++

// As changes the base currency
func (data FromFixer) As(name string) FromFixer {
	if data.Base == name {
		return data
	}
	var baseCurrency float32
	if data.Base == "EUR" {
		baseCurrency = 1.0
	} else {
		baseCurrency = data.Rates[data.Base]
	}
	data.Rates[data.Base] = 1 * GetRates(data.Rates[name], baseCurrency)

	data.Base = name
	baseValue := data.Rates[name]
	for key, value := range data.Rates {
		data.Rates[key] = 1 * GetRates(baseValue, value)
	}
	delete(data.Rates, name)

	return data
}

//++++++++++++++++++++++++++ getRates +++++++++++++++++++++++++++++++++++++

// GetRates returns the currency rates
func GetRates(from float32, to float32) float32 {
	if from == to {
		return 1.0
	}
	return to * (1 / from)
}

//+++++++++++++++++++++++++ TO ++++++++++++++++++++++++++++++++++++++++++++

// To returns the value from a currency to another
func (data FromFixer) To(name string) Convertion {
	return Convertion{
		From:      data.Base,
		FromValue: 1.0,
		To:        name,
		ToValue:   data.Rates[name],
		Rate:      data.Rates[name],
	}
}

//----------------------------------------------------------------------------
func handlerpost(res http.ResponseWriter, req *http.Request) {

	var webHook WebHook
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&webHook)
	if err != nil {
		res.WriteHeader(200)
		return
	}

	webHook.ID = bson.NewObjectId()
	mongoTickets.add(webHook)
	//Returne response code
	res.WriteHeader(http.StatusCreated)
	fmt.Fprintln(res, webHook.ID.Hex())

}

//----------------------------------------------------------------------------
func handlerEx(res http.ResponseWriter, req *http.Request) {
	ting := mux.Vars(req)
	if !bson.IsObjectIdHex(ting["ID"]) {
		res.WriteHeader(400)
		fmt.Fprintf(res, "Internal error")
		return
	}
	webshit := mongoTickets.get(ting["ID"])
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, webshit)
}

//---------------------------------------------------------------------------
func handlerDel(res http.ResponseWriter, req *http.Request) {
	ting := mux.Vars(req)
	if !bson.IsObjectIdHex(ting["ID"]) {
		res.WriteHeader(400)
		fmt.Fprintf(res, "Internal error")
		return
	}
	mongoTickets.delete(ting["ID"])
	res.WriteHeader(200)
}

//---------------------------------------------------------------------------
func handlerAver(res http.ResponseWriter, req *http.Request) {
	var webhook LatestRates
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&webhook)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprint(res, aver(&webhook))
}

//---------------------------------------------------------------------------

func handlerlate(res http.ResponseWriter, req *http.Request) {
	var webhook LatestRates
	var js FromDialog
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&js)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	webhook.BaseCurrency = js.Result.Parameters.BaseCurrency
	webhook.TargetCurrency = js.Result.Parameters.TargetCurrency
	//just for testing purposes
	fmt.Print(js.Result.Parameters.BaseCurrency)
	fmt.Print(js.Result.Parameters.TargetCurrency)
	//----- -----   ----   ----- ----   ----
	fmt.Fprint(res, latest(&webhook))
}
