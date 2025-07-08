package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	port := flag.String("port", "3000", "Port to run the server on")
	p := flag.String("p", "", "Shorthand for --port")
	folder := flag.String("folder", "./", "Folder to serve static files from")
	f := flag.String("f", "", "Shorthand for --folder")

	flag.Parse()

	if *p != "" {
		*port = *p
	}
	if *f != "" {
		*folder = *f
	}

	address := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting server on %s\n", address)
	fmt.Printf("Serving files from %s\n", *folder)

	router := http.NewServeMux()

	router.Handle("/", http.FileServer(http.Dir(*folder)))

	handler := cors.AllowAll().Handler(router)

	err := http.ListenAndServe(address, handler)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
