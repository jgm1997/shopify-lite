package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func RespondWithJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return err
	}
	return nil
}

func ValidateEmail(email string) bool {
	return len(email) > 3 && len(email) < 254 && strings.Contains(email, "@")
}

func ValidatePassword(password string) bool {
	return len(password) >= 8
}

func ToInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float32:
		return int(t)
	case float64:
		return int(t)
	case []byte:
		n, err := strconv.Atoi(string(t))
		if err == nil {
			return n
		}
	case string:
		n, err := strconv.Atoi(t)
		if err == nil {
			return n
		}
	}
	return 0
}

func ToFloat64(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case []byte:
		n, err := strconv.ParseFloat(string(t), 64)
		if err == nil {
			return n
		}
	case string:
		n, err := strconv.ParseFloat(t, 64)
		if err == nil {
			return n
		}
	}
	return 0
}
