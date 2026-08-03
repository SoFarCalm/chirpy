package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	auth "github.com/SoFarCalm/chirpy/internal/auth"
	"github.com/SoFarCalm/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ChirpyRed    bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerUpgradeUserStatus(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid API key", err)
		return
	}

	if apiKey != cfg.PolkaKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key", nil)
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			ID uuid.UUID `json:"user_id"`
		}
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Invalid event type", nil)
		log.Printf("Invalid event type: %s", params.Event)
		return
	}

	if params.Event == "user.upgraded" {
		log.Printf("Upgrading user status for user ID: %s", params.Data.ID)
		_, err := cfg.db.UpdateUserChirpyRedStatus(r.Context(), database.UpdateUserChirpyRedStatusParams{
			ID:          params.Data.ID,
			IsChirpyRed: sql.NullBool{Bool: true, Valid: true},
		})
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Couldn't update user status", err)
			return
		}
	}

	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
	}

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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Fatal(err)
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             id,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			ChirpyRed: user.IsChirpyRed.Bool,
		},
	})

}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Fatal(err)
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			ChirpyRed: user.IsChirpyRed.Bool,
		},
	})
}
