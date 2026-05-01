package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spoddub/subscription-aggregator/internal/model"
)

const (
	invalidIDError          string = "invalid id"
	invalidRequestBodyError string = "invalid request body"
)

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          int64     `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *Handler) ListSubscriptions(c *gin.Context) {
	subscriptions, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list subscriptions",
		})
		return
	}

	response := make([]SubscriptionResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		response = append(response, toSubscriptionResponse(subscription))
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": response,
	})
}

func (h *Handler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": invalidRequestBodyError,
		})
		return
	}

	params, err := buildCreateSubscriptionParams(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	subscription, err := h.repo.Create(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create subscription",
		})
		return
	}

	c.JSON(http.StatusCreated, toSubscriptionResponse(subscription))
}

func (h *Handler) GetSubscriptionByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": invalidIDError,
		})
		return
	}

	subscription, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get subscription",
		})
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(subscription))
}

func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": invalidIDError,
		})
		return
	}

	var req UpdateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": invalidRequestBodyError,
		})
		return
	}

	params, err := buildUpdateSubscriptionParams(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	subscription, err := h.repo.Update(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(subscription))
}

func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": invalidIDError,
		})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete subscription",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetTotalCost(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	userIDRaw := c.Query("user_id")
	serviceNameRaw := strings.TrimSpace(c.Query("service_name"))

	fromDate, err := parseMonthYear(from)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid from, expected format MM-YYYY",
		})
		return
	}

	toDate, err := parseMonthYear(to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid to, expected format MM-YYYY",
		})
		return
	}

	if fromDate.After(toDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be before or equal to to",
		})
		return
	}

	var userID *uuid.UUID
	if userIDRaw != "" {
		parsedUserID, err := uuid.Parse(userIDRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": invalidIDError,
			})
			return
		}

		userID = &parsedUserID
	}

	var serviceName *string
	if serviceNameRaw != "" {
		serviceName = &serviceNameRaw
	}

	total, err := h.repo.GetTotalCost(c.Request.Context(), model.TotalCostFilter{
		From:        fromDate,
		To:          toDate,
		UserID:      userID,
		ServiceName: serviceName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to calculate total cost",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from":         from,
		"to":           to,
		"user_id":      userIDRaw,
		"service_name": serviceNameRaw,
		"total":        total,
	})
}

func buildCreateSubscriptionParams(req CreateSubscriptionRequest) (model.CreateSubscriptionParams, error) {
	serviceName := strings.TrimSpace(req.ServiceName)
	if serviceName == "" {
		return model.CreateSubscriptionParams{}, errors.New("service_name is required")
	}

	if req.Price <= 0 {
		return model.CreateSubscriptionParams{}, errors.New("price must be greater than zero")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return model.CreateSubscriptionParams{}, errors.New(invalidIDError)
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		return model.CreateSubscriptionParams{}, errors.New("invalid start_date, expected format MM-YYYY")
	}

	endDate, err := parseOptionalMonthYear(req.EndDate)
	if err != nil {
		return model.CreateSubscriptionParams{}, errors.New("invalid end_date, expected format MM-YYYY")
	}

	if endDate != nil && endDate.Before(startDate) {
		return model.CreateSubscriptionParams{}, errors.New("end_date must be after or equal to start_date")
	}

	return model.CreateSubscriptionParams{
		ServiceName: serviceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

func buildUpdateSubscriptionParams(
	id int64,
	req UpdateSubscriptionRequest,
) (model.UpdateSubscriptionParams, error) {
	serviceName := strings.TrimSpace(req.ServiceName)
	if serviceName == "" {
		return model.UpdateSubscriptionParams{}, errors.New("service_name is required")
	}

	if req.Price <= 0 {
		return model.UpdateSubscriptionParams{}, errors.New("price must be greater than zero")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return model.UpdateSubscriptionParams{}, errors.New(invalidIDError)
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		return model.UpdateSubscriptionParams{}, errors.New("invalid start_date, expected format MM-YYYY")
	}

	endDate, err := parseOptionalMonthYear(req.EndDate)
	if err != nil {
		return model.UpdateSubscriptionParams{}, errors.New("invalid end_date, expected format MM-YYYY")
	}

	if endDate != nil && endDate.Before(startDate) {
		return model.UpdateSubscriptionParams{}, errors.New("end_date must be after or equal to start_date")
	}

	return model.UpdateSubscriptionParams{
		ID:          id,
		ServiceName: serviceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

func parseID(rawID string) (int64, error) {
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, errors.New("id must be greater than zero")
	}

	return id, nil
}

func parseMonthYear(value string) (time.Time, error) {
	return time.Parse("01-2006", strings.TrimSpace(value))
}

func parseOptionalMonthYear(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := parseMonthYear(value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func formatMonthYear(value time.Time) string {
	return value.Format("01-2006")
}

func toSubscriptionResponse(subscription model.Subscription) SubscriptionResponse {
	var endDate *string
	if subscription.EndDate != nil {
		formattedEndDate := formatMonthYear(*subscription.EndDate)
		endDate = &formattedEndDate
	}

	return SubscriptionResponse{
		ID:          subscription.ID,
		ServiceName: subscription.ServiceName,
		Price:       subscription.Price,
		UserID:      subscription.UserID,
		StartDate:   formatMonthYear(subscription.StartDate),
		EndDate:     endDate,
		CreatedAt:   subscription.CreatedAt,
		UpdatedAt:   subscription.UpdatedAt,
	}
}
