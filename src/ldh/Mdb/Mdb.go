package Mdb

import (
	"gopkg.in/mgo.v2"
	"log"
)

var (
	session      *mgo.Session
	databaseName = "CoolpyV"
	DatabaseAddress = ""
)
func Session() *mgo.Session {
	if session == nil {
		var err error
		session, err = mgo.Dial(DatabaseAddress)
		if err != nil {
			log.Fatal(err) // no, not really
		}
	}
	return session.Clone()
}
func M(collection string, f func(*mgo.Collection)) {
	session := Session()
	defer func() {
		session.Close()
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	c := session.DB(databaseName).C(collection)
	f(c)
}
