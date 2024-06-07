package main

import (
	"net/http"

	"github.com/gothew/hogger"
)

func main() {
	http.Handle("/", hogger.Middleware(http.HandlerFunc(handler)))
	go func() {
		http.ListenAndServe(":1337", nil)
	}()

	h := "http://localhost:1337"
	c := &http.Client{}
	c.Get(h + "/")
	r, _ := http.NewRequest("POST", h+"/meow", nil)
	c.Do(r)
	r, _ = http.NewRequest("PUT", h+"/purr", nil)
	c.Do(r)
	c.Get(h + "/schnurr")
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.RequestURI {
	case "/":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("oh bea"))

	case "/meow":
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte("over"))
	case "/purr":
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not here"))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("wrong"))
	}
}
