package users

import (
	"encoding/json"
	"net/http"
	"strings"

	"shopify-lite/internal/auth"
	"shopify-lite/internal/middleware"
)

func respondWithJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return err
	}
	return nil
}

func validateEmail(email string) bool {
	// Basic email validation (can be improved with regex)
	return len(email) > 3 && len(email) < 254 && strings.Contains(email, "@")
}

func validatePassword(password string) bool {
	return len(password) >= 8
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var u User
	if err := decodeJSONBody(w, r, &u); err != nil {
		return
	}
	if !validateEmail(u.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	if !validatePassword(u.Password) {
		http.Error(w, "password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	created, err := h.users.CreateUser(r.Context(), u)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	token, err := auth.GenerateToken(created.ID, string(created.Role), h.jwtSecret)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	respondWithJson(w, http.StatusCreated, map[string]any{
		"token": token,
	})
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var u struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(w, r, &u); err != nil {
		return
	}
	if u.Email == "" || u.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}
	email := u.Email
	password := u.Password

	user, found, err := h.users.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}
	if !found || !auth.CheckPassword(user.Password, password) {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateToken(user.ID, string(user.Role), h.jwtSecret)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	respondWithJson(w, http.StatusOK, map[string]any{
		"token": token,
	})
}
func (h *Handler) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r) // ✅ trust the middleware
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	u, found, err := h.users.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to fetch user", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	respondWithJson(w, http.StatusOK, u.ToResponse()) // ✅ no password
}
