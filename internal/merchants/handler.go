package merchants

import (
	"net/http"
	"shopify-lite/internal/middleware"
	"shopify-lite/internal/utils"
)

func (h *Handler) CreateMerchantHandler(w http.ResponseWriter, r *http.Request) {
	var m Merchant
	if err := utils.DecodeJSONBody(w, r, &m); err != nil {
		return
	}

	created, err := h.merchants.CreateMerchant(r.Context(), m)
	if err != nil {
		http.Error(w, "failed to create merchant", http.StatusInternalServerError)
		return
	}
	utils.RespondWithJson(w, http.StatusCreated, created)
}

func (h *Handler) GetMeMerchantHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	me, found, err := h.merchants.GetMerchantByUserID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to fetch merchant", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "merchant not found", http.StatusNotFound)
		return
	}
	utils.RespondWithJson(w, http.StatusOK, me)
}
