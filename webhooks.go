package main

import (
	"encoding/json"
	"net/http"

	"github.com/MaazU-Dev/chirpy/internal/auth"
	"github.com/MaazU-Dev/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "API Key not provided", err)
		return
	}
	if key != cfg.polkaApiKey {
		respondWithError(w, http.StatusUnauthorized, "Incorrect API Key", err)
		return
	}
	type data struct {
		UserId string `json:"user_id"`
	}
	type reqBody struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}
	var request reqBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}
	if request.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	parsedId, err := uuid.Parse(request.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}
	err = cfg.dbQueries.UpdateUserToChirpyRed(r.Context(), database.UpdateUserToChirpyRedParams{
		IsChirpyRed: true,
		ID:          parsedId,
	})
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
