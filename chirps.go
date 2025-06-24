package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()
	bannedWords := os.Getenv("BANNED_WORDS")

	type Chirp struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Chirp{}
	err := decoder.Decode(&params)
	if err != nil || len(params.Body) == 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		log.Printf("Error: decoding chirp: %v", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp body exceeds 140 characters")
		log.Printf("Error: Chirp body exceeds 140 characters: %s...", params.Body[:140])
		return
	}

	words := strings.Split(params.Body, " ")
	bannedWordsList := strings.Split(bannedWords, " ")
	for i, word := range words {
		for _, bannedWord := range bannedWordsList {
			if strings.EqualFold(word, bannedWord) {
				words[i] = "****"
				break
			}
		}
	}
	cleanedChirp := strings.Join(words, " ")

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}
	respondWithJSON(w, http.StatusOK, response{CleanedBody: cleanedChirp})
	log.Printf("Received chirp: %s", params.Body)
}