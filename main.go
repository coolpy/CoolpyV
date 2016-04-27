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
	"ldh/Mdb"
)

func main() {
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
		panic(err)
		return
	}
	json.NewEncoder(w).Encode(p)
}

  func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	  Mdb.M("people", func(c *mgo.Collection) {
		  result := []Models.Person{}
		  err := c.Find(bson.M{"name": "Ale"}).All(&result)
		  if err != nil {
			  log.Fatal(err)
		  }
		  json.NewEncoder(w).Encode(result)
	  })
  }

  func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
      fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
  }
