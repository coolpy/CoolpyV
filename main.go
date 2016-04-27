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
)
type Person struct {
	Name string
	Phone string
}
func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Basicauth.Auth(Hello))
	err := http.ListenAndServe(":8080", Cors.CORS(router))
	if err != nil {
		log.Fatal(err)
	}

	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("test").C("people")
	err = c.Insert(&Person{"Ale", "+55 53 8116 9639"},
		&Person{"Cla", "+55 53 8402 8510"})
	if err != nil {
		log.Fatal(err)
	}

	result := Person{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Phone:", result.Phone)
}
  func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
      fmt.Fprint(w, "Welcome!\n")
  }

  func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
      fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
  }
