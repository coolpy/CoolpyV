package main

import (
	"log"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"ldh/Cors"
	"ldh/BasicAuth"
	"ldh/Models"
	"encoding/json"
)

func main() {
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("people")
	err = c.Insert(&Models.Person{"Ale", "+55 53 8116 9639"},&Models.Person{"Cla", "+55 53 8402 8510"})
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/",IndexPost)
	router.GET("/hello/:name", Basicauth.Auth(Hello))

	log.Fatal(http.ListenAndServe(":8080", Cors.CORS(router)))
}

func IndexPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params){
	decoder := json.NewDecoder(r.Body)
	var p Models.Person
	err := decoder.Decode(&p)
	if err != nil {
		panic("error")
		return
	}
	json.NewEncoder(w).Encode(p)
}

  func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	  session, err := mgo.Dial("127.0.0.1:27017")
	  if err != nil {
		  panic(err)
	  }
	  defer session.Close()
	  // Optional. Switch the session to a monotonic behavior.
	  session.SetMode(mgo.Monotonic, true)
	  c := session.DB("test").C("people")
	  result := Models.Person{}
	  err = c.Find(bson.M{"name": "Ale"}).One(&result)
	  if err != nil {
		  log.Fatal(err)
	  }
	  json.NewEncoder(w).Encode(result)
  }

  func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
      fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
  }
