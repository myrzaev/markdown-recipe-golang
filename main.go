package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

//go:embed index.html
var html string

const (
	pocketBaseURL = "http://127.0.0.1:8090"
	serverPort    = ":8080"
)

var sessions = make(map[string]string)

func main() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		return se.Next()
	})

	go func() {
		if err := app.Start(); err != nil {
			log.Fatalf("Failed to start PocketBase: %v", err)
		}
	}()

	component := hello("John")
	// http.HandleFunc("/", serveIndex)
	// http.HandleFunc("/login", serveLogin)
	http.Handle("/", templ.Handler(component))
	http.HandleFunc("/auth", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	// http.HandleFunc("/api/recipes", proxyRequest("recipes"))
	// http.HandleFunc("/api/ratings", proxyRequest("ratings"))

	mux := setupRoutes()

	fmt.Printf("Server is running on http://localhost%s\n", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, mux))
}

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/api/recipes", authMiddleware(proxyRequest("recipes")))
	mux.Handle("/api/ratings", authMiddleware(proxyRequest("ratings")))

	return mux
}

// func serveIndex(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 	w.Write([]byte(html))
// }

// func serveIndex(w http.ResponseWriter, r *http.Request) {
// 	component := hello("John")
// 	component.Render(context.Background(), os.Stdout)
// }

// func serveLogin(w http.ResponseWriter, r *http.Request) {
// 	component := hello("John")
// 	component.Render(context.Background(), os.Stdout)
// }

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	url := fmt.Sprintf("%s/api/collections/users/auth-with-password", pocketBaseURL)
	requestBody, _ := json.Marshal(map[string]string{
		"identity": username,
		"password": password,
	})

	resp, err := http.Post(url, "application/json", io.NopCloser(bytes.NewReader(requestBody)))
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	defer resp.Body.Close()

	sessionID := fmt.Sprintf("sess-%d", time.Now().Unix())
	sessions[sessionID] = username

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   "",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, &cookie)
	w.Write([]byte("Logged out"))
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || sessions[cookie.Value] == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func proxyRequest(collection string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("%s/api/collections/%s/records", pocketBaseURL, collection)
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Failed to connect to PocketBase", http.StatusInternalServerError)
			log.Printf("Error fetching %s: %v", collection, err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			log.Printf("Error reading response body for %s: %v", collection, err)
			return
		}

		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			http.Error(w, "Failed to parse response", http.StatusInternalServerError)
			log.Printf("Error parsing JSON response for %s: %v", collection, err)
			return
		}

		if items, ok := responseData["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					delete(itemMap, "collectionId")
					delete(itemMap, "collectionName")
				}
			}
		}

		cleanedJSON, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Printf("Error encoding cleaned JSON response for %s: %v", collection, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(cleanedJSON)
	}
}

// router, database connection, pocketbase, templ (template language, no REACT, server side render)
// login page, cookies (cookie base, session base) authentication, alpine jS

// 1 pocketbase
// 2 router
// 3 login support (cookie)
