package main

import (
	"github.com/dthxsu/Chirpy/ChirpyProject/internal/database"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"path/filepath"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	dbPath, err := filepath.Abs(filepath.Join(filepathRoot, "internal", "database", "database.json"))
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the database
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
	}

	router := chi.NewRouter()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	router.Handle("/app", fsHandler)
	router.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
	apiRouter.Post("/chirps", apiCfg.handlerChirpsCreate)
	apiRouter.Get("/chirps", apiCfg.handlerChirpsRetrieve)
	apiRouter.Get("/chirps/{id}", apiCfg.handlerChirpRetrieveByID)
	apiRouter.Post("/users", apiCfg.handlerUsersCreate)
	apiRouter.Post("/login", apiCfg.handlerLogin)
	router.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)
	router.Mount("/admin", adminRouter)

	corsMux := middlewareCors(router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
