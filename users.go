package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leonardomlouzas/GOose/internal/auth"
	"github.com/leonardomlouzas/GOose/internal/database"
)

type User struct {
	ID        		uuid.UUID	`json:"id"`
	Email     		string	    `json:"email"`
	CreatedAt 		time.Time	`json:"created_at"`
	UpdatedAt 		time.Time	`json:"updated_at"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email		string	`json:"email"`
		Password	string	`json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		log.Printf("error decoding request payload while creating user: %v", err)
		return
	}

	email := strings.TrimSpace(params.Email)
	password := strings.TrimSpace(params.Password)

	if email == "" || password == "" {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		log.Print("error while creating user. Email/Password not provided")
		return
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid password")
		log.Printf("error while hashing password. Error: %s", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        		uuid.New(),
		Email:     		email,
		CreatedAt: 		time.Now().UTC(),
		UpdatedAt: 		time.Now().UTC(),
		HashedPassword:	hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating user")
		log.Printf("error inserting user into db while creating user: %s", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (cfg *apiConfig) handlerGetAllUsers(w http.ResponseWriter, r *http.Request) {
	dbUsers, err := cfg.db.GetAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error retrieving users")
		log.Printf("error retrieving users table: %s", err)
		return
	}

	users := make([]User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = User{
			ID:        dbUser.ID,
			Email:     dbUser.Email,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		}
	}
	respondWithJSON(w, http.StatusOK, users)
}

func (cfg *apiConfig) handlerGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid user ID")
		log.Printf("error parsing user ID: %s while retrieving user by ID. Error: %s", userID, err)
		return
	}

	user, err := cfg.db.GetUserById(r.Context(), uid)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "user not found")
			return
		}
	
		respondWithError(w, http.StatusInternalServerError, "error retrieving user")
		log.Printf("error retrieving user by id: %s. Error: %v", userID, err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (cfg *apiConfig) handlerLoginByPassword(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email		string	`json:"email"`
		Password	string	`json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		log.Printf("error decoding request payload while logging in: %v", err)
		return
	}

	email := strings.TrimSpace(params.Email)
	password := strings.TrimSpace(params.Password)

	if email == "" || password == "" {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		log.Print("error while login user. Email/Password not provided")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect Email/Password")
		log.Printf("error while retrieving user by email while login. Error: %s", err)
		return
	}

	if err := auth.CheckPasswordHash(password, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect Email/Password")
		return
	}


	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}