package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/MaazU-Dev/chirpy/internal/auth"
	"github.com/MaazU-Dev/chirpy/internal/database"
	"github.com/google/uuid"
)

// func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Add("Content-Type", "application/json")

// 	type parameters struct {
// 		Body string `json:"body"`
// 	}
// 	type resBody struct {
// 		CleanedBody string `json:"cleaned_body"`
// 	}
// 	var val parameters
// 	decoder := json.NewDecoder(r.Body)
// 	if err := decoder.Decode(&val); err != nil {
// 		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
// 		return
// 	}

// 	const maxChirpLength = 140
// 	if len(val.Body) > maxChirpLength {
// 		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
// 		return
// 	}

// 	cleaned := profaneFilter(val.Body)

// 	respondWithJSON(w, http.StatusOK, resBody{
// 		CleanedBody: cleaned,
// 	})
// }

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

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type resBody struct {
		Chirp
	}
	var reqBody parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	const maxChirpLength = 140
	if len(reqBody.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanedBody := profaneFilter(reqBody.Body)

	jwt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Unable to get bearer token", err)
		return
	}

	userId, err := auth.ValidateJWT(jwt, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Unable to validate JWT", err)
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create UUID", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, resBody{
		Chirp: Chirp{
			ID:        chirp.ID,
			UserID:    chirp.UserID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		},
	})
}

func (cfg *apiConfig) handleChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	data, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to get all chirps", err)
		return
	}
	listChirps := make([]Chirp, len(data))
	for i, val := range data {
		listChirps[i] = Chirp{
			ID:        val.ID,
			UserID:    val.UserID,
			Body:      val.Body,
			CreatedAt: val.CreatedAt,
			UpdatedAt: val.UpdatedAt,
		}
	}
	respondWithJSON(w, http.StatusOK,
		listChirps,
	)
}

func (cfg *apiConfig) handleChirpsRetrieveByID(w http.ResponseWriter, r *http.Request) {
	type resBody struct {
		Chirp
	}
	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Id is required", nil)
		return
	}
	parsedId, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Id is not in correct format", err)
		return
	}
	data, err := cfg.dbQueries.GetChirpsByID(r.Context(), parsedId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to get all chirps", err)
		return
	}
	respondWithJSON(w, http.StatusOK,
		resBody{
			Chirp: Chirp{
				ID:        data.ID,
				UserID:    data.UserID,
				Body:      data.Body,
				CreatedAt: data.CreatedAt,
				UpdatedAt: data.UpdatedAt,
			},
		},
	)
}

func (cfg *apiConfig) HandleChirpsDeleteByID(w http.ResponseWriter, r *http.Request) {
	jwt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Unable to get bearer token", err)
		return
	}
	userId, err := auth.ValidateJWT(jwt, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Unable to validate JWT", err)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Id is required", nil)
		return
	}
	parsedId, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Id is not in correct format", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirpsByID(r.Context(), parsedId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to get chirp", err)
		return
	}
	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "Unauthorized: You are not the owner of this chirp", nil)
		return
	}
	err = cfg.dbQueries.DeleteChirpsByID(r.Context(), parsedId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to found the chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
