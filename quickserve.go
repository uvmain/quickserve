package main

import (
	"crypto/subtle"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type filteredFileSystem struct {
	root http.FileSystem
}

func (fs filteredFileSystem) Open(name string) (http.File, error) {
	if isPasswordPath(name) {
		return nil, os.ErrNotExist
	}

	f, err := fs.root.Open(name)
	if err != nil {
		return nil, err
	}

	return filteredFile{File: f}, nil
}

type filteredFile struct {
	http.File
}

func (f filteredFile) Readdir(count int) ([]os.FileInfo, error) {
	if count <= 0 {
		entries, err := f.File.Readdir(count)
		if err != nil {
			return nil, err
		}
		return filterPasswordEntries(entries), nil
	}

	result := make([]os.FileInfo, 0, count)
	for len(result) < count {
		entries, err := f.File.Readdir(count - len(result))
		if len(entries) > 0 {
			result = append(result, filterPasswordEntries(entries)...)
		}

		if err != nil {
			if err == io.EOF {
				if len(result) > 0 {
					return result, nil
				}
				return nil, io.EOF
			}
			return result, err
		}

		if len(entries) == 0 {
			break
		}
	}

	if len(result) == 0 {
		return nil, io.EOF
	}

	if len(result) > count {
		result = result[:count]
	}

	return result, nil
}

func filterPasswordEntries(entries []os.FileInfo) []os.FileInfo {
	filtered := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if !strings.EqualFold(entry.Name(), ".password") {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func isPasswordPath(name string) bool {
	clean := path.Clean("/" + name)
	return strings.EqualFold(path.Base(clean), ".password")
}

func isWithinBaseDir(candidatePath, baseDir string) bool {
	rel, err := filepath.Rel(baseDir, candidatePath)
	if err != nil {
		return false
	}

	if rel == "." {
		return true
	}

	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}

	return !filepath.IsAbs(rel)
}

func handlerFunc(baseDir string) http.Handler {
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		fmt.Printf("Error resolving base directory: %s\n", err)
		os.Exit(1)
	}
	absBaseDir = filepath.Clean(absBaseDir)
	fileServer := http.FileServer(filteredFileSystem{root: http.Dir(absBaseDir)})

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
		absPath, ok := resolveSecurePath(requestPath, absBaseDir)
		if !ok {
			fmt.Printf("Blocked path traversal attempt: %s from %s\n", requestPath, r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if strings.EqualFold(filepath.Base(absPath), ".password") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if !checkDirectoryPasswordAuth(w, r, absBaseDir, absPath) {
			return
		}

		timestamp := time.Now().Format(time.RFC3339Nano)
		fmt.Printf("%s %s %s from %s\n", timestamp, r.Method, r.RequestURI, r.RemoteAddr)

		fileServer.ServeHTTP(w, r)
	})
}

func resolveSecurePath(requestPath, baseDir string) (string, bool) {
	if strings.Contains(requestPath, "\x00") {
		return "", false
	}

	cleanPath := path.Clean("/" + requestPath)
	if strings.Contains(cleanPath, "..") {
		return "", false
	}

	relPath := strings.TrimPrefix(cleanPath, "/")
	fullPath := filepath.Join(baseDir, filepath.FromSlash(relPath))
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", false
	}

	if !isWithinBaseDir(absPath, baseDir) {
		return "", false
	}

	return absPath, true
}

func checkDirectoryPasswordAuth(w http.ResponseWriter, r *http.Request, baseDir, absPath string) bool {
	targetDir := absPath
	if info, err := os.Stat(absPath); err == nil {
		if !info.IsDir() {
			targetDir = filepath.Dir(absPath)
		}
	} else {
		targetDir = filepath.Dir(absPath)
	}

	password, protectedDir, ok := findDirectoryPassword(targetDir, baseDir)
	if !ok {
		return true
	}

	_, providedPassword, authOK := r.BasicAuth()
	if authOK && subtle.ConstantTimeCompare([]byte(providedPassword), []byte(password)) == 1 {
		return true
	}

	realm := filepath.ToSlash(strings.TrimPrefix(protectedDir, baseDir))
	if realm == "" {
		realm = "/"
	}
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"Protected %s\"", realm))
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return false
}

func findDirectoryPassword(startDir, baseDir string) (string, string, bool) {
	for dir := startDir; ; dir = filepath.Dir(dir) {
		if !isWithinBaseDir(dir, baseDir) {
			return "", "", false
		}

		passwordPath := filepath.Join(dir, ".password")
		data, err := os.ReadFile(passwordPath)
		if err == nil {
			password := strings.TrimSpace(string(data))
			return password, dir, true
		}

		if dir == baseDir {
			break
		}
	}

	return "", "", false
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
