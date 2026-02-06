package main

import (
	"net/http"
	"time"

	"github.com/MaazU-Dev/chirpy/internal/auth"
)

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	type resBody struct {
		AccessToken string `json:"token"`
	}
	refreskToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token not provided", err)
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreskToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Refresh token not found", err)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to create JWt", err)
		return
	}

	respondWithJSON(w, http.StatusOK, resBody{
		AccessToken: jwt,
	})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreskToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token not provided", err)
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreskToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Write header is used to send status code to the client
}
