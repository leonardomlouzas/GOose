package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leonardomlouzas/GOose/internal/database"
)

type Chirp struct {
	ID			uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Body		string		`json:"body"`
	UserID		uuid.UUID	`json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type ChirpReq struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := ChirpReq{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cleanedBody, err := validateAndCleanChirp(params.Body, cfg.bannedWords)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	user, err := cfg.db.GetUserById(r.Context(), params.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		log.Printf("Error: user not found with ID %s while posting a chirp: %v", params.UserID, err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      cleanedBody,
		UserID:    user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
		log.Printf("Error: creating chirp: %v", err)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	log.Printf("Chirp %s created by user %s", chirp.ID, chirp.UserID)
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps")
		log.Printf("Error retrieving chirps: %s", err)
	}

	respondWithJSON(w, http.StatusOK, chirps)
	log.Printf("Chirps retrieved successfully")
}

func (cfg *apiConfig) handlerGetOneChirp(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	chirpID := params.Get("id")

	uid, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		log.Printf("Error parsing chirp id: %s. Error: %s", chirpID, err)
	}

	chirp, err := cfg.db.GetOneChirp(r.Context(), uid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("getting chirp by id: %v", err))
		log.Printf("Error retrieving chirp by id: %s", err)
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}

func validateAndCleanChirp(body string, bannedWords map[string]struct{}) (string, error) {
	if len(body) == 0 {
		return "", fmt.Errorf("chirp body cannot be empty")
	}
	if len(body) > 140 {
		return "", fmt.Errorf("chirp is too long")
	}

	words := strings.Split(body, " ")
	for i, word := range words {
		if _, ok := bannedWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " "), nil
}