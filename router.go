package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

type ActionRequest struct {
	ID     string    `json:"id"`
	Action Operation `json:"action,omitempty"`
}

type Operation struct {
	Name   string `json:"name"`
	Serial string `json:"serial,omitempty"`
}

func NewActionRequest(serial string, name string) *ActionRequest {
	op := Operation{
		Name:   name,
		Serial: serial,
	}

	id := uuid.Must(uuid.NewV4())
	return &ActionRequest{
		ID:     id.String(),
		Action: op,
	}
}

func GetActionRequest(w http.ResponseWriter, r *http.Request) {
	action := NewActionRequest("testSerial", "rebootDevice")
	log.Printf("Request Action #%v \n", action.ID)
	json.NewEncoder(w).Encode(action)
}

func DeleteActionRequest(w http.ResponseWriter, r *http.Request) {
	log.Println(httputil.DumpRequestOut(r, true))
	params := mux.Vars(r)
	id := params["ID"]
	log.Printf("Action #%v can be removed\n", id)
	w.WriteHeader(http.StatusOK)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/action/{id}", GetActionRequest).Methods("GET")
	router.HandleFunc("/action/{id}", DeleteActionRequest).Methods("DELETE")
	port := "8000"

	srv := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%v", port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      logRequest(router), // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
