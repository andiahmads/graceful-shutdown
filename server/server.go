package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/v1/readiness", readiness)

	server := http.Server{
		Addr:    "0.0.0.0:3001",
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Println("server listening on", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	// ListenAndServe dijalankan di goroutine yang lain agar line berikutnya dapat dieksekusi,
	// karena kita tahu bahwa ListenAndServe mengakibatkan blocking.
	// Perubahan yang kedua adalah sekarang kita sudah memiliki mekanisme untuk mengamati sinyal yang dapat mengakibatkan shutdown.

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT)

	select {
	case sig := <-shutdownChannel:
		log.Println("signal:", sig)

		timeout := 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
		}
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("server: %s", err)
		}
	}

}

func readiness(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-REQUEST-ID")
	log.Println("start", requestID)
	defer log.Println("done", requestID)

	time.Sleep(5 * time.Second)
	response := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		panic(err)
	}
}
