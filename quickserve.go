package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func handlerFunc(baseDir string) http.Handler {
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		fmt.Printf("Error resolving base directory: %s\n", err)
		os.Exit(1)
	}
	absBaseDir = filepath.Clean(absBaseDir)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		requestPath := r.URL.Path
		if !isSecurePath(requestPath, absBaseDir) {
			fmt.Printf("Blocked path traversal attempt: %s from %s\n", requestPath, r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		timestamp := time.Now().Format(time.RFC3339Nano)
		fmt.Printf("%s %s %s from %s\n", timestamp, r.Method, r.RequestURI, r.RemoteAddr)

		http.FileServer(http.Dir(absBaseDir)).ServeHTTP(w, r)
	})
}

func isSecurePath(requestPath, baseDir string) bool {
	cleanPath := filepath.Clean(requestPath)

	// check for path traversal || null byte
	if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "\x00") {
		return false
	}

	fullPath := filepath.Join(baseDir, cleanPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false
	}

	// Ensure the resolved path is still within the base directory
	return strings.HasPrefix(absPath, baseDir+string(filepath.Separator)) || absPath == baseDir
}

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
	fmt.Printf("Starting secure server on %s\n", address)
	fmt.Printf("Serving files from %s\n", *folder)

	handler := handlerFunc(*folder)

	err := http.ListenAndServe(address, handler)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
