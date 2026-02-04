package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	type parameters struct {
		Body string `json:"body"`
	}
	type resBody struct {
		CleanedBody string `json:"cleaned_body"`
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

	cleaned := profaneFilter(val.Body)

	respondWithJSON(w, http.StatusOK, resBody{
		CleanedBody: cleaned,
	})
}

func profaneFilter(body string) string {
	list := strings.Split(body, " ")
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	for i, val := range list {
		lowerVal := strings.ToLower(val)
		if _, ok := badWords[lowerVal]; ok {
			list[i] = "****"
		}
	}

	return strings.Join(list, " ")
}
