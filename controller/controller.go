// Package controller This is swagger for Gallery
//
// The purpose of this application is to retrieve and store images
//
//
//
//     BasePath: /api/v1
//     Version: 1.0.0
//     Contact: Mohit Arora<mohitarora19966@gmail.com>
//
//     Consumes:
//       - application/json
//
//     Produces:
//       - application/json
//
//
// swagger:meta
package controller

//go:generate swagger generate spec -m -o ./swagger.json

import (
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	//ErrBadRequest for bad request errors
	ErrBadRequest = errors.New("bad reqeust")
)

func swaggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "swagger.json")
}

// RunController is used to run server in given port
func RunController(host string) {
	r := mux.NewRouter()

	// album handler
	r.HandleFunc("/api/v1/albums", albumPostHandler).Methods("POST")
	r.HandleFunc("/api/v1/albums/{albumID}", albumDeleteHandler).Methods("DELETE")
	r.HandleFunc("/api/v1/albums", albumGetHandler).Methods("GET")

	// image handler
	r.HandleFunc("/api/v1/albums/{albumID}/images", imagePostHandler).Methods("POST")
	r.HandleFunc("/api/v1/albums/{albumID}/images/{imageID}", imageDeleteHandler).Methods("DELETE")
	r.HandleFunc("/api/v1/albums/{albumID}/images", imageGetHandler).Methods("GET")

	r.HandleFunc("/swagger", swaggerHandler)

	fs := http.FileServer(http.Dir("./swaggerui"))
	r.PathPrefix("/swaggerui/").Handler(http.StripPrefix("/swaggerui/", fs))

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("/data/"))))

	log.Println("starting server at ", host)
	err := http.ListenAndServe(host, r)
	if err != nil {
		panic(err)
	}
}
