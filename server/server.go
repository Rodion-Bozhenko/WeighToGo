package server

import (
	"net/http"
)

func Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SUCK DIZ NUTS"))
	})

	http.ListenAndServe("localhost:5000", nil)
}
