package orders

import (
	"errors"
	"net/http"
	"shopify-lite/internal/middleware"
	"shopify-lite/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetCustomerOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	orders, err := h.orders.GetCustomerOrders(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "failed to get customer orders", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, orders)
}

func (h *Handler) GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil || orderID <= 0 {
		http.Error(w, "invalid order ID", http.StatusBadRequest)
		return
	}

	order, found, err := h.orders.GetOrderByID(r.Context(), orderID, claims.UserID)
	if err != nil {
		http.Error(w, "failed to get order", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	utils.RespondWithJson(w, http.StatusOK, order)
}

func (h *Handler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req PlaceOrderRequest
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, ErrEmptyItems.Error(), http.StatusBadRequest)
		return
	}

	order, ok, err := h.orders.PlaceOrder(r.Context(), claims.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, ErrInsufficientStock):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "failed to place order", http.StatusInternalServerError)
		}
		return
	}
	if !ok {
		http.Error(w, "failed to place order", http.StatusInternalServerError)
		return
	}
	utils.RespondWithJson(w, http.StatusCreated, order)
}
