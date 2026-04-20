package products

import (
	"context"
	"net/http"
	"shopify-lite/internal/middleware"
	"shopify-lite/internal/utils"
)

const failedFetchMerchant = "failed to fetch merchant"

func (h *Handler) getMerchantByUserID(ctx context.Context, userID int) (int, error) {
	return h.products.GetMerchantIDByUserID(ctx, userID)
}

func (h *Handler) GetMyProductsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	merchantID, err := h.getMerchantByUserID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, failedFetchMerchant, http.StatusInternalServerError)
		return
	}

	products, err := h.products.GetMyProducts(r.Context(), merchantID)
	if err != nil {
		http.Error(w, "failed to fetch products", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, products)
}

func (h *Handler) CreateMyProductHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var p Product
	if err := utils.DecodeJSONBody(w, r, &p); err != nil {
		return
	}

	merchantID, err := h.products.GetMerchantIDByUserID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, failedFetchMerchant, http.StatusInternalServerError)
		return
	}

	p.MerchantID = merchantID
	created, err := h.products.CreateMyProduct(r.Context(), p)
	if err != nil {
		http.Error(w, "failed to create product", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, created)
}

func (h *Handler) UpdateMyProductHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var p Product
	if err := utils.DecodeJSONBody(w, r, &p); err != nil {
		return
	}

	merchantID, err := h.products.GetMerchantIDByUserID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, failedFetchMerchant, http.StatusInternalServerError)
		return
	}

	p.MerchantID = merchantID
	updated, err := h.products.UpdateMyProduct(r.Context(), p)
	if err != nil {
		http.Error(w, "failed to update product", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, updated)
}

func (h *Handler) DeleteMyProductHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var p Product
	if err := utils.DecodeJSONBody(w, r, &p); err != nil {
		return
	}

	merchantID, err := h.products.GetMerchantIDByUserID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, failedFetchMerchant, http.StatusInternalServerError)
		return
	}

	p.MerchantID = merchantID
	deleted, err := h.products.DeleteMyProduct(r.Context(), p.ID)
	if err != nil {
		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, map[string]bool{"deleted": deleted})
}

func (h *Handler) GetMyProductsDashboardHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	merchantID, err := h.products.GetMerchantIDByUserID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, failedFetchMerchant, http.StatusInternalServerError)
		return
	}

	dashboard, err := h.products.GetMyProductsDashboard(r.Context(), merchantID)
	if err != nil {
		http.Error(w, "failed to fetch dashboard", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, dashboard)
}
