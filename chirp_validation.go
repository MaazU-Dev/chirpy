package main

import (
	"encoding/json"
	"net/http"
)

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	type parameters struct {
		Body string `json:"body"`
	}
	type resBody struct {
		Valid bool `json:"valid"`
	}
	var val parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&val); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	const maxChirpLength = 140
	if len(val.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	respondWithJSON(w, http.StatusOK, resBody{
		Valid: true,
	})
}
