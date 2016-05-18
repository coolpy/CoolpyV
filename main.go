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
	"os"
)

var configuration interface{}
var Version = "5.0.0.0"

func main() {
	fmt.Printf("Coolpy Version: %s\n", Version)
	if _, err := os.Stat("conf.json"); err != nil {
		log.Fatal("Config file is missing: ", "conf.json")
		os.Exit(1)
	}
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configuration); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	conf, ok := configuration.(map[string]interface{})
	if ok {
		for k, v := range conf {
			switch v2 := v.(type) {
			case string:
				if k == "mongo" {
					Mdb.DatabaseAddress = v2
				}
			default:
				fmt.Println(k, v)
			}
		}
	}
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/", IndexPost)
	router.GET("/hello/:name", Basicauth.Auth(Hello))
	if err := http.ListenAndServe(":8080", Cors.CORS(router)); err != nil {
		log.Fatal(err)
	}
}

func IndexPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var p Models.Person
	err := decoder.Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Fatal(err)
		return
	}
	Mdb.M("people", func(c *mgo.Collection) {
		err = c.Insert(p)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			log.Fatal(err)
			return
		}
		json.NewEncoder(w).Encode(p)
	})
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
