package main

import (
	"encoding/json"
	"net/http"
	"time"

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
		Email string `json:"email"`
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

	user, err := cfg.dbQueries.CreateUser(r.Context(), request.Email)
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
