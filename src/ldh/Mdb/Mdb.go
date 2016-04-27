package Mdb

import (
	"gopkg.in/mgo.v2"
	"log"
)

var (
	session      *mgo.Session
	databaseName = "test"
)
func Session() *mgo.Session {
	if session == nil {
		var err error
		session, err = mgo.Dial("127.0.0.1:27017")
		if err != nil {
			panic(err) // no, not really
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
