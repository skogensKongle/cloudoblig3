package main

import (
  "testing"
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


func Testrates_getRates(t *testing.T){
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
