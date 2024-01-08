package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type apiConfig struct {
	fileserverHits int
}

type Chirp struct {
	Body string `json:"body"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := &apiConfig{}
	r := chi.NewRouter()

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", http.HandlerFunc(healthzHandler))
	apiRouter.Get("/metrics", http.HandlerFunc(apiCfg.metricsHandler))
	apiRouter.Get("/reset", http.HandlerFunc(apiCfg.resetHandler))
	apiRouter.Post("/validate_chirp", http.HandlerFunc(validateChirpHandler))
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", http.HandlerFunc(apiCfg.metricsHTMLHandler))
	r.Mount("/admin", adminRouter)

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) metricsHTMLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	htmlTemplate := `<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>`
	w.Write([]byte(fmt.Sprintf(htmlTemplate, cfg.fileserverHits)))
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func cleanChirp(chirp string) string {
	profanities := []string{"kerfuffle", "sharbert", "fornax"}
	for _, profanity := range profanities {
		chirp = strings.Replace(chirp, profanity, "****", -1)
		chirp = strings.Replace(chirp, strings.Title(profanity), "****", -1)

	}
	return chirp
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var chirp Chirp
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(chirp.Body) > 140 {
		http.Error(w, `{"error": "Chirp is too long"}`, http.StatusBadRequest)
		return
	}

	cleanedChirp := cleanChirp(chirp.Body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"cleaned_body": "` + cleanedChirp + `"}`))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " + strconv.Itoa(cfg.fileserverHits)))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
}
