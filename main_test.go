package main

import (
	"fmt"
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

func TestMongo_get(t *testing.T) {
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)

	//nr := db.Count()
	if db.Count() != 0 {
		t.Error("database not properly initialized, data count() should be ", db.Count())
	}

	data := WebHook{Webhookurl: "testytest.org", Basecurrency: "TES", Targetcurrency: "SET", Mintriggervalue: 1.234, Maxtriggervalue: 2.543}
	data.ID = bson.NewObjectId()
	db.add(data)

	if db.Count() != 1 {
		t.Error("adding new webhook failed.", db.Count())
	}
	var newData WebHook
	objectid := bson.ObjectId(data.ID).Hex()
	//uncomment the line under to check what the objectid is
	//fmt.Print(objectid)
	newData = db.get(objectid)

	if newData.Webhookurl != data.Webhookurl ||
		newData.Basecurrency != data.Basecurrency ||
		newData.Targetcurrency != data.Targetcurrency ||
		newData.Mintriggervalue != data.Mintriggervalue ||
		newData.Maxtriggervalue != data.Maxtriggervalue {
		t.Error("data do not match.", newData.ID, newData.Webhookurl, newData.Basecurrency, newData.Targetcurrency, newData.Mintriggervalue, newData.Maxtriggervalue)
	}
}

func TestDel_delete(t *testing.T) {
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)

	if db.Count() != 0 {
		t.Error("database not properly initialized, data count() should be 0.", db.Count())
	}

	data := WebHook{Webhookurl: "testytest.org", Basecurrency: "TES", Targetcurrency: "SET", Mintriggervalue: 1.234, Maxtriggervalue: 2.543}
	data.ID = bson.NewObjectId()
	db.add(data)
	objectid := bson.ObjectId(data.ID).Hex()
	//fmt.Print(objectid)

	if db.Count() != 1 {
		t.Error("adding new webhook failed.", db.Count())
	}

	db.delete(objectid)
	if db.Count() != 0 {
		t.Error("Could not delete.", db.Count())
	}
}

func Testaverage_aver(t *testing.T) {
	db := Mongo{DatabaseURL: "mongodb://localhost", DatabaseName: "testing", MongoCollection: "test"}
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()
	if err != nil {
		t.Error(err)
	}
	defer tearDownDB(t, &db)

	if db.Count() != 0 {
		t.Error("database not properly initialized, data count() should be 0.", db.Count())
	}

	data := FromFixer{Base: "EUR", Date: "2017-11-10", Rates: map[string]float32{"NOK": 1}}
	err = session.DB(db.DatabaseName).C(db.MongoCollection).Insert(data)
	if err != nil {
		t.Error(err)
	}
	data2 := FromFixer{Base: "EUR", Date: "2017-11-11", Rates: map[string]float32{"NOK": 1}}
	err = session.DB(db.DatabaseName).C(db.MongoCollection).Insert(data2)
	if err != nil {
		t.Error(err)
	}

	if db.Count() != 2 {
		t.Error("adding new webhook failed.", db.Count())
	}

	var test LatestRates
	test.BaseCurrency = "EUR"
	test.TargetCurrency = "NOK"

	if aver(&test) != 1 {
		t.Error("Did not finde average")
	}
}
