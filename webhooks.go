package main

import (
	"encoding/json"
	"net/http"

	"github.com/MaazU-Dev/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
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
