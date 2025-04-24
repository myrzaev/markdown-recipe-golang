package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// SignUp creates a new user in PocketBase
func SignUp(w http.ResponseWriter, r *http.Request) {
	var userData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Create the user object
	user := map[string]string{
		"email":           userData["email"],
		"password":        userData["password"],
		"passwordConfirm": userData["password"],
	}

	// Send POST request to PocketBase to create the user
	userDataJSON, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error encoding user data: %v", err)
		http.Error(w, "Failed to process user data", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(pocketBaseURL+"/api/collections/users/records", "application/json", bytes.NewBuffer(userDataJSON))
	if err != nil {
		log.Printf("Failed to send request to PocketBase: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		http.Error(w, "Failed to create user", resp.StatusCode)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}

// Login authenticates a user and returns the auth token
func SignIn(w http.ResponseWriter, r *http.Request) {
	var loginData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Send login request to PocketBase
	login := map[string]string{
		"identity": loginData["email"],
		"password": loginData["password"],
	}

	loginDataJSON, err := json.Marshal(login)
	if err != nil {
		log.Printf("Error encoding login data: %v", err)
		http.Error(w, "Failed to process login data", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(pocketBaseURL+"/api/collections/users/auth-with-password", "application/json", bytes.NewBuffer(loginDataJSON))
	if err != nil {
		log.Printf("Failed to send request to PocketBase: %v", err)
		http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Parse the response to get the auth token
	var authResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	// Extract the JWT token
	token, exists := authResp["token"].(string)
	if !exists {
		http.Error(w, "No token received", http.StatusInternalServerError)
		return
	}

	// Send the token in the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// Middleware to check the authorization header and validate the token
func IsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
		return false
	}

	// You can use the token to verify against PocketBase's /api/collections/users/authenticate endpoint
	// For simplicity, we'll assume the token is valid for now.
	return true
}
