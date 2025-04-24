package handlers

import (
	"io"
	"net/http"
)

func GetRatings(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(pocketBaseURL + "/api/collections/ratings/records?expand=user")
	if err != nil {
		http.Error(w, "Failed to fetch ratings", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch ratings from PocketBase", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
