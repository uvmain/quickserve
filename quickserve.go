package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	port := flag.String("port", "3000", "Port to run the server on")
	p := flag.String("p", "", "Shorthand for port")

	flag.Parse()

	if *p != "" {
		*port = *p
	}

	address := fmt.Sprintf("localhost:%s", *port)
	fmt.Printf("Starting server on %s\n", address)
	http.Handle("/", http.FileServer(http.Dir("./")))
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
