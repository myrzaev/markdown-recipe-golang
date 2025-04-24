package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/myrzaev/markdown-recipe-golang/handlers"
)

const (
	pocketBaseURL = "http://127.0.0.1:8090"
	serverPort    = ":8080"
)

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running 🚀")
	})

	http.HandleFunc("/api/auth/sign-up", handlers.SignUp)
	http.HandleFunc("/api/auth/sign-in", handlers.SignIn)
	http.HandleFunc("/api/auth/verification", Verification)
	http.HandleFunc("/api/recipes", handlers.GetRecipes)
	http.HandleFunc("/api/ratings", handlers.GetRatings)
	http.HandleFunc("/api/ratings/average/", handlers.GetAverageRecipeRating)

	fmt.Printf("Server is running on http://localhost%s\n", serverPort)
	err := http.ListenAndServe(serverPort, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// A protected route that requires authentication
func Verification(w http.ResponseWriter, r *http.Request) {
	if !handlers.IsAuthorized(w, r) {
		return
	}

	// Continue with the protected logic
	w.WriteHeader(http.StatusOK)
}
