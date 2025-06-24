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

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("decoding request payload while creating user: %v", err))
		return
	}

	email := strings.TrimSpace(params.Email)
	if email == "" {
		respondWithError(w, http.StatusBadRequest, "email not provided while creating user")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		Email:     email,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("creating user: %v\n", err))
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
	log.Printf("User ID %s created: %s\n",user.ID, user.Email)
}