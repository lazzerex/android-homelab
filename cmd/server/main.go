package main

import (
	"log"
	"net/http"
	"os"

	"android-go-server-test/internal"
)

func main() {
	port:=os.Getenv("PORT")
	if port=="" { port="8080" }

	http.HandleFunc("/", internal.IndexHandler)
	http.HandleFunc("/health", internal.HealthHandler)
	http.HandleFunc("/info", internal.InfoHandler)
	http.HandleFunc("/time", internal.TimeHandler)

	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port,nil))
}
