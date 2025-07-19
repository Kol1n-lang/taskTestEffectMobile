package handler

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"taskTestEffectMobile/internal/models/json_models"
	"taskTestEffectMobile/internal/service"
	"taskTestEffectMobile/internal/utils"
)

type SubscriptionHandler struct {
	service  service.SubscriptionService
	validate *validator.Validate
	logger   *zap.Logger
}

func NewSubscriptionHandler(s service.SubscriptionService, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service:  s,
		validate: validator.New(),
		logger:   logger,
	}
}

func (subscriptionHandler *SubscriptionHandler) CreateSubscriptionsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/subscriptions/create-subscription", subscriptionHandler.createSubscription)
	mux.HandleFunc("GET /api/v1/subscriptions/get-subscription", subscriptionHandler.getSubscription)
	mux.HandleFunc("PUT /api/v1/subscriptions/update-subscription", subscriptionHandler.updateSubscription)
	mux.HandleFunc("DELETE /api/v1/subscriptions/delete-subscription", subscriptionHandler.deleteSubscription)
	mux.HandleFunc("GET /api/v1/subscriptions/calculate-cost", subscriptionHandler.calculateSubscriptionsCost)
}

// createSubscription creates a new subscription
// @Summary Create subscription
// @Description Creates a new subscription for user
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscription body json_models.CreateSubscription true "Subscription data"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Failed to decode JSON request"
// @Failure 422 {string} string "Validation error"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/create-subscription [post]
func (subscriptionHandler *SubscriptionHandler) createSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionHandler.logger.Info("Create subscription request received")

	var subscription json_models.CreateSubscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		subscriptionHandler.logger.Error("Failed to decode JSON request",
			zap.Error(err),
			zap.String("path", r.URL.Path))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := subscriptionHandler.validate.Struct(subscription); err != nil {
		subscriptionHandler.logger.Warn("Validation error",
			zap.Error(err),
			zap.Any("subscription", subscription))
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	subscriptionHandler.logger.Info("Creating subscription",
		zap.String("userID", subscription.UserID),
		zap.String("serviceName", subscription.ServiceName))

	userUUID, err := subscriptionHandler.service.CreateSubscription(r.Context(), subscription)
	if err != nil {
		subscriptionHandler.logger.Error("Failed to create subscription",
			zap.Error(err),
			zap.Any("subscription", subscription))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if userUUID == "" {
		subscriptionHandler.logger.Warn("Subscription already exists",
			zap.String("userID", subscription.UserID),
			zap.String("serviceName", subscription.ServiceName))
		http.Error(w, "Subscription already exists", http.StatusConflict)
		return
	}

	response := map[string]string{
		"id":     userUUID,
		"status": "created",
	}

	subscriptionHandler.logger.Info("Subscription created successfully",
		zap.String("subscriptionID", userUUID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		subscriptionHandler.logger.Error("Failed to encode response",
			zap.Error(err),
			zap.Any("response", response))
	}
}

// getSubscription retrieves user's subscriptions
// @Summary Get subscriptions
// @Description Returns all subscriptions for specified user
// @Tags Subscriptions
// @Produce json
// @Param user-id query string true "User ID"
// @Success 200 {array} sql_models.Subscription "List of subscriptions"
// @Failure 400 {string} string "Invalid UUID format"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/get-subscription [get]
func (subscriptionHandler *SubscriptionHandler) getSubscription(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	userID := params.Get("user-id")

	subscriptionHandler.logger.Info("Get subscription request",
		zap.String("userID", userID))

	if userID == "" {
		subscriptionHandler.logger.Warn("Missing user-id parameter")
		http.Error(w, "Missing user-id", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		subscriptionHandler.logger.Warn("Invalid UUID format",
			zap.String("userID", userID),
			zap.Error(err))
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	response, err := subscriptionHandler.service.GetUserSubscriptions(r.Context(), userUUID)
	if err != nil {
		subscriptionHandler.logger.Error("Failed to get subscriptions",
			zap.String("userID", userID),
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	subscriptionHandler.logger.Info("Successfully retrieved subscriptions",
		zap.String("userID", userID),
		zap.Int("subscriptionCount", len(response)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		subscriptionHandler.logger.Error("Failed to encode response",
			zap.Error(err),
			zap.Any("response", response))
	}
}

// updateSubscription updates subscription data
// @Summary Update subscription
// @Description Updates existing subscription data
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscription body json_models.PutSubscription true "Update data"
// @Success 202 {object} map[string]string
// @Failure 400 {string} string "Invalid request format"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/update-subscription [put]
func (subscriptionHandler *SubscriptionHandler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionHandler.logger.Info("Update subscription request received")

	var subscription json_models.PutSubscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		subscriptionHandler.logger.Error("Failed to decode JSON request",
			zap.Error(err))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := subscriptionHandler.validate.Struct(subscription); err != nil {
		subscriptionHandler.logger.Warn("Validation failed",
			zap.Error(err),
			zap.Any("subscription", subscription))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	subscriptionHandler.logger.Info("Updating subscription",
		zap.String("subscriptionID", subscription.SubscriptionID),
		zap.String("serviceName", subscription.ServiceName))

	err := subscriptionHandler.service.UpdateSubscription(r.Context(), subscription)
	if err != nil {
		subscriptionHandler.logger.Error("Failed to update subscription",
			zap.Error(err),
			zap.Any("subscription", subscription))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	subscriptionHandler.logger.Info("Subscription updated successfully",
		zap.String("subscriptionID", subscription.SubscriptionID),
		zap.String("serviceName", subscription.ServiceName))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	response := map[string]string{
		"status": "updated",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		subscriptionHandler.logger.Error("Failed to encode response",
			zap.Error(err))
	}
}

// deleteSubscription removes a subscription
// @Summary Delete subscription
// @Description Deletes specified subscription
// @Tags Subscriptions
// @Param subscription-id query string true "Subscription ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Invalid parameters"
// @Failure 404 {string} string "Subscription not found"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/delete-subscription [delete]
func (subscriptionHandler *SubscriptionHandler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionIDStr := r.URL.Query().Get("subscription-id")

	subscriptionHandler.logger.Info("Delete subscription request",
		zap.String("userID", subscriptionIDStr))

	if subscriptionIDStr == "" {
		subscriptionHandler.logger.Warn("Invalid query parameters",
			zap.String("userID", subscriptionIDStr))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	subscriptionUUID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		subscriptionHandler.logger.Warn("Invalid subscription ID format",
			zap.String("userID", subscriptionIDStr),
			zap.Error(err))
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	if err := subscriptionHandler.service.DeleteSubscription(r.Context(), subscriptionUUID); err != nil {
		subscriptionHandler.logger.Error("Failed to delete subscription",
			zap.String("userID", subscriptionIDStr),
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	subscriptionHandler.logger.Info("Subscription deleted successfully",
		zap.String("userID", subscriptionIDStr))

	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"status": "deleted",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		subscriptionHandler.logger.Error("Failed to encode response",
			zap.Error(err))
	}
}

// calculateSubscriptionsCost calculates total cost of subscriptions
// @Summary Calculate subscriptions cost
// @Description Calculates total cost of subscriptions for given period with optional filters
// @Tags Subscriptions
// @Param user-id query string false "User ID filter"
// @Param service-name query string false "Service name filter"
// @Param start-date query string true "Start date (format: 01-2006)"
// @Param end-date query string false "End date (format: 01-2006)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Invalid query parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/calculate-cost [get]
func (subscriptionHandler *SubscriptionHandler) calculateSubscriptionsCost(w http.ResponseWriter, r *http.Request) {
	subscriptionHandler.logger.Info("Handling subscriptions cost calculation request")

	var req json_models.CostRequest
	if err := utils.QueryParser(r, &req); err != nil {
		subscriptionHandler.logger.Error("Failed to parse query params", zap.Error(err))
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	if err := subscriptionHandler.validate.Struct(req); err != nil {
		subscriptionHandler.logger.Warn("Validation failed",
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID *uuid.UUID
	if req.UserID != nil {
		id, err := uuid.Parse(*req.UserID)
		if err != nil {
			subscriptionHandler.logger.Warn("Invalid user ID format",
				zap.String("userID", *req.UserID),
				zap.Error(err))
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}
		userID = &id
	}

	totalCost, err := subscriptionHandler.service.CalculateSubscriptionsCost(
		r.Context(),
		userID,
		req.ServiceName,
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		subscriptionHandler.logger.Error("Failed to calculate subscriptions cost",
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if req.EndDate != nil {
		response := map[string]interface{}{
			"period": map[string]string{
				"start": req.StartDate,
				"end":   *req.EndDate,
			},
			"total_cost": totalCost,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			subscriptionHandler.logger.Error("Failed to encode response",
				zap.Error(err))
		}
		return
	}
	response := map[string]interface{}{
		"total_cost": totalCost,
		"period": map[string]string{
			"start": req.StartDate,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		subscriptionHandler.logger.Error("Failed to encode response",
			zap.Error(err))
	}
}
