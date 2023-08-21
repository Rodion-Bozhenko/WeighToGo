package server

import (
	"fmt"
	"log"
	"net/http"
)

func Server(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(fmt.Sprintf("Response from :%v\n", addr)))
		if err != nil {
			log.Printf("Failed to write response on %s: %v", addr, err)
		}
	})

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalf("Failed to start server on %s: %v", addr, err)
	}
}
