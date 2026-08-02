package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/SoFarCalm/chirpy/internal/auth"
	"github.com/SoFarCalm/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	reqChirpID := r.PathValue("chirpID")
	parsedID, err := uuid.Parse(reqChirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find Chirp :( ", err)
		return
	}

	if chirp.UserID != id {
		respondWithError(w, http.StatusForbidden, "You are not the owner of this chirp", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {
	reqChirpID := r.PathValue("chirpID")
	log.Printf("This is the requested id: %s", reqChirpID)

	parsedID, err := uuid.Parse(reqChirpID)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}

	chirp, err := cfg.db.GetChirp(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find Chirp :( ", err)
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {

	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		return
	}

	var chirpsReturned []Chirp

	for _, chirp := range chirps {
		chirpsReturned = append(chirpsReturned, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirpsReturned)

}

func (cfg *apiConfig) handlerChirpCreate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid token", err)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
		//UserID uuid.UUID `json:"user_id"`
	}
	type returnVals struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	params.Body = replaceBadWords(params.Body, []string{"kerfuffle", "sharbert", "fornax"})

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: id,
	})
	if err != nil {
		log.Printf("Error creating chirp %s", err)
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
