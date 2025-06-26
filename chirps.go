package main

import (
	"database/sql"
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
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		log.Printf("error decoding request payload while posting chirp: %v", err)
		return
	}

	cleanedBody, err := validateAndCleanChirp(params.Body, cfg.bannedWords)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	user, err := cfg.db.GetUserById(r.Context(), params.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user not found")
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
		respondWithError(w, http.StatusInternalServerError, "error creating chirp")
		log.Printf("error inserting chirp into db while posting chirp: %v", err)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error retrieving chirps")
		log.Printf("error retrieving chirps table: %s", err)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetOneChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("id")
	uid, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		log.Printf("error parsing chirp ID: %s while retrieving chirp by id. Error: %s", chirpID, err)
		return
	}

	chirp, err := cfg.db.GetOneChirp(r.Context(), uid)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "chirp not found")
			return
		}
	
		respondWithError(w, http.StatusInternalServerError, "error retrieving chirp by id")
		log.Printf("error retrieving chirp by id %s: %v", chirpID, err)
		return
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