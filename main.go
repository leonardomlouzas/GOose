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

const filepathRoot = "."
const port = "8080"

type apiConfig struct {
	db				*database.Queries
	fileserverHits	atomic.Int32
	env				string
	bannedWords		map[string]struct{}
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
	if cfg.env != "dev" {
		respondWithError(w, http.StatusForbidden, "Not allowed in production environment")
		return
	}

	cfg.fileserverHits.Store(0)
	cfg.db.ResetUsersTable(r.Context())

	respondWithJSON(w, http.StatusOK, "Reset successful")
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	env := os.Getenv("ENVIRONMENT")
	bannedWordsRaw := os.Getenv("BANNED_WORDS")
	bannedWordsList := strings.Split(bannedWordsRaw, " ")
	bannedWordsMap := make(map[string]struct{})
	for _, word := range bannedWordsList {
		bannedWordsMap[strings.ToLower(word)] = struct{}{}
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v\n", err)
	}
	defer db.Close()

	apiCfg := &apiConfig{
		db: database.New(db),
		fileserverHits: atomic.Int32{},
		env:            env,
		bannedWords:    bannedWordsMap,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("GET /api/users", apiCfg.handlerGetUserByID)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChirp)
	
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	log.Printf("Responding with %d error: %s", code, msg)
	type errResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", payload)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}
