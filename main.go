package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/leonardomlouzas/GOose/internal/database"
	_ "github.com/lib/pq"
)

const bannedWords = "kerfuffle sharbert fornax"
const filepathRoot = "."
const port = "8080"

type apiConfig struct {
	db				*database.Queries
	fileserverHits	atomic.Int32
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>
	`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body 	string		`json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Chirp{}
	err := decoder.Decode(&params)
	if err != nil || len(params.Body) == 0{
		response := map[string]string{
			"error": "Invalid request body",
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseJSON)
		log.Printf("Error: decoding chirp: %v", err)
		return
	}

	if len(params.Body) > 140 {
		response := map[string]string{
			"error": "Chirp body exceeds 140 characters",
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseJSON)
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"cleaned_body": cleanedChirp,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(responseJSON)
	log.Printf("Received chirp: %s", params.Body)
}

func main() {
	godotenv.Load()
	
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v\n", err)
	}
	defer db.Close()

	apiCfg := &apiConfig{
		db: database.New(db),
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerPostChirp)
	
	server := &http.Server {
		Addr:		":" + port,
		Handler: 	mux,
	}
	log.Printf("Serving files from %s under /app/ on port: %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", server.Addr, err)
	}
	
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
