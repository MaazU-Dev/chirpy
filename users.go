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
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
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
			ID:          user.ID,
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			IsChirpyRed: user.IsChirpyRed,
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
			ID:          user.ID,
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        jwt,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, http.StatusOK, res)
}

func (cfg *apiConfig) handleUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type resBody struct {
		User
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "You are not authorized", err)
		return
	}
	_, err = auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "You are not authorized", err)
		return
	}
	/// Fetch the user record(Combinde both queries and use a transaction, SHOULD WE?) from the user ID received from the JWT
	// Run validation on email and password
	// Email should be same, password should not be same
	var request reqBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Provide email or password", err)
		return
	}
	hash, err := auth.HashPassword(request.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create hash", err)
		return
	}
	updatedUser, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          request.Email,
		HashedPassword: hash,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update user", err)
		return
	}
	respondWithJSON(w, http.StatusOK, resBody{
		User: User{
			ID:          updatedUser.ID,
			Email:       updatedUser.Email,
			CreatedAt:   updatedUser.CreatedAt,
			UpdatedAt:   updatedUser.UpdatedAt,
			IsChirpyRed: updatedUser.IsChirpyRed,
		},
	})
}
