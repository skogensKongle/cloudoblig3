package main

import (
	"fmt"
	"testing"

	"gopkg.in/mgo.v2"
)

var testRates = FromFixer{Base: "EUR", Date: "2017-11-10", Rates: map[string]float32{"BGN": 1.9558000564575195, "USD": 1.1654000282287598, "SEK": 9.743000030517578, "NOK": 9.456000328063965}}

func tearDownDB(t *testing.T, db *Mongo) {
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}

	err = session.DB(db.DatabaseName).DropDatabase()
	if err != nil {
		t.Error(err)
	}
}

func (db *Mongo) Count() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.MongoCollection).Count()
	if err != nil {
		fmt.Printf("error in Count(): %v", err.Error())
		return -1
	}
	return count
}

func TestMongo_add(t *testing.T) {
	//db := setupDB(t)
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)

	//nr := db.Count()
	data := WebHook{Webhookurl: "testytest.org", Basecurrency: "TES", Targetcurrency: "SET", Mintriggervalue: 1.234, Maxtriggervalue: 2.543}
	db.add(data)

	if db.Count() != 1 {
		t.Error("adding new webhook failed.")
	}
}

func Testrates_getRates(t *testing.T) {
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)
	/*
	     db.add(data)

	   	if db.Count() != 1 {
	   		t.Error("adding new webhook failed.")
	   	}
	     ++TO DO++
	     -need to get it out and then check if it matches.
	*/
}
