package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const pocketBaseURL = "http://127.0.0.1:8090"

func GetRecipes(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(pocketBaseURL + "/api/collections/recipes/records?expand=author")
	if err != nil {
		http.Error(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch recipes from PocketBase", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func GetAverageRecipeRating(w http.ResponseWriter, r *http.Request) {
	recipeID := strings.TrimPrefix(r.URL.Path, "/api/ratings/average/") // Retrieve recipe ID from URL
	log.Printf("Fetching average rating for recipe ID: %s", recipeID)

	// Request ratings for a specific recipe
	resp, err := http.Get(pocketBaseURL + "/api/collections/ratings/records?filter=recipeId='" + recipeID + "'")
	if err != nil {
		log.Printf("Failed to fetch ratings: %v", err)
		http.Error(w, "Failed to fetch ratings", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch ratings from PocketBase. Status code: %d", resp.StatusCode) // logging status code error

		http.Error(w, "Failed to fetch ratings from PocketBase", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// Parsing JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
		return
	}

	// Retrieve scores from the “items” field
	items := result["items"].([]interface{})

	// Calculation of average rating
	var totalScore int
	for _, item := range items {
		rating := item.(map[string]interface{})
		score, err := strconv.Atoi(fmt.Sprintf("%v", rating["value"]))
		if err != nil {
			http.Error(w, "Invalid score value", http.StatusInternalServerError)
			return
		}
		totalScore += score
	}

	// If there are no ratings, return 0
	if len(items) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"averageRating": 0}`))
		return
	}

	averageRating := float64(totalScore) / float64(len(items))

	// Send average rating
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]float64{
		"averageRating": averageRating,
	})
}
