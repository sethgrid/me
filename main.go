package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

func init() {
	defaultPort := 6000
	envPort := os.Getenv("MAIN_PORT")
	if envPort != "" {
		var err error
		defaultPort, err = strconv.Atoi(envPort)
		if err != nil {
			log.Fatal("provide MAIN_PORT as an integer", envPort, err)
		}
	}

	flag.IntVar(&Port, "port", defaultPort, "default port 6000")
}

var (
	Port int
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/posts/{name}", PostsHandler).Methods("GET", "POST")
	http.Handle("/", r)

	log.Printf("running service on :%d with resources\n\t/\n\t/posts/:post_name", Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Port), nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("home"))
}

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	c, err := redis.Dial("tcp", "192.168.59.103:6379")
	if err != nil {
		log.Println("error dialing redis")
		return
	}
	defer c.Close()
	key := vars["name"]

	reply, err := c.Do("GET", key)
	if err != nil {
		handleErr(w, "error with GET", key, err)
		return
	}

	var value int64
	switch reply.(type) {
	case nil:
		log.Println("new key", key)
	case []byte:
		str := string(reply.([]byte))
		i64, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			handleErr(w, "error converting to int ", reply, str, err)
			return
		}
		value = i64
	default:
		_, _ = c.Do("DEL", key)
		log.Println("removing key, non recognized type for value ", reply)
	}

	value++

	_, err = c.Do("SET", key, value)
	if err != nil {
		handleErr(w, "error with SET ", key, err)
		return
	}

	response := fmt.Sprintf("you requested `%s` [%d] times", key, value)
	w.Write([]byte(response))
}

func handleErr(w http.ResponseWriter, i ...interface{}) {
	w.Write([]byte(fmt.Sprint(i...)))
}
