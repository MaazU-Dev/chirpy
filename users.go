package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MaazU-Dev/chirpy/internal/auth"
	"github.com/MaazU-Dev/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *apiConfig) handleUsersCreate(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type resBody struct {
		User
	}
	var request reqBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to hash the password", err)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          request.Email,
		HashedPassword: hash,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create User", err)
		return
	}
	res := resBody{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}
	respondWithJSON(w, http.StatusCreated, res)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type resBody struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	var request reqBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong", err)
		return
	}

	user, err := cfg.dbQueries.GetUser(r.Context(), request.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User does not exists", err)
		return
	}

	same, err := auth.CheckPasswordHash(request.Password, user.HashedPassword)
	if err != nil || !same {
		respondWithError(w, http.StatusUnauthorized, "Password does not match", err)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create Auth Token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create Refresh Token", err)
		return
	}

	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 30),
		UserID:    user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create Refresh Token in db", err)
		return
	}

	res := resBody{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token:        jwt,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, res)
}
