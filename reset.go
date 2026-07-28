package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Forbidden Action", errors.New("Forbidden"))
	}

	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		fmt.Println("Error deleting users")
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
