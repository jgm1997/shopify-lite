package products

import (
	"net/http"
	"shopify-lite/internal/middleware"
	"shopify-lite/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const failedFetchMerchant = "failed to fetch merchant"

func (h *Handler) GetMyProductsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	products, err := h.products.GetMyProducts(r.Context(), claims.MerchantID)
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

	p.MerchantID = claims.MerchantID
	created, err := h.products.CreateMyProduct(r.Context(), p)
	if err != nil {
		http.Error(w, "failed to create product", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusCreated, created)
}

func (h *Handler) UpdateMyProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var p Product
	if err := utils.DecodeJSONBody(w, r, &p); err != nil {
		return
	}
	p.ID = id
	p.MerchantID = claims.MerchantID

	updated, found, err := h.products.UpdateMyProduct(r.Context(), p)
	if err != nil {
		http.Error(w, "failed to update product", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "product not found or not owned by merchant", http.StatusNotFound)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, updated)
}

func (h *Handler) DeleteMyProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deleted, err := h.products.DeleteMyProduct(r.Context(), id, claims.MerchantID)
	if err != nil {
		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "product not found or not owned by merchant", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMyProductsDashboardHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	dashboard, err := h.products.GetMyProductsDashboard(r.Context(), claims.MerchantID)
	if err != nil {
		http.Error(w, "failed to fetch dashboard", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, dashboard)
}
