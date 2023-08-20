package server

import (
	"fmt"
	"net/http"
)

func Server(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("Response from :%v\n", addr)))
	})

	http.ListenAndServe(addr, mux)
}
