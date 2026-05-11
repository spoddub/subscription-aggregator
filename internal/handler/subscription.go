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
	invalidUserIDError      string = "invalid user_id"
	invalidRequestBodyError string = "invalid request body"
	subscriptionNotFound    string = "subscription not found"
	defaultListLimit               = 20
	maxListLimit                   = 100
	defaultOffset                  = 0
)

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int    `json:"price" example:"400"`
	UserID      string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string `json:"start_date" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" example:"12-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" example:"Yandex Plus"`
	Price       *int    `json:"price,omitempty" example:"500"`
	UserID      *string `json:"user_id,omitempty" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   *string `json:"start_date,omitempty" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

type SubscriptionResponse struct {
	ID          int64     `json:"id" example:"1"`
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int       `json:"price" example:"400"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
	Limit         int                    `json:"limit" example:"20"`
	Offset        int                    `json:"offset" example:"0"`
	Count         int                    `json:"count" example:"1"`
}

type TotalCostResponse struct {
	From        string `json:"from" example:"07-2025"`
	To          string `json:"to" example:"09-2025"`
	UserID      string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Total       int64  `json:"total" example:"1200"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request body"`
}

// ListSubscriptions godoc
// @Summary List subscriptions
// @Description Returns subscriptions. Supports pagination and optional filters by user_id and service_name.
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Limit, default 20, max 100"
// @Param offset query int false "Offset, default 0"
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Subscription service name"
// @Success 200 {object} ListSubscriptionsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	filter, err := buildListSubscripionsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriptions, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to list subscriptions",
			"error", err,
			"limit", filter.Limit,
			"offset", filter.Offset)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list subscriptions",
		})
		return
	}

	response := make([]SubscriptionResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		response = append(response, toSubscriptionResponse(subscription))
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"subscriptions listed",
		"count", len(response),
		"limit", filter.Limit,
		"offset", filter.Offset)

	c.JSON(http.StatusOK, ListSubscriptionsResponse{
		Subscriptions: response,
		Limit:         filter.Limit,
		Offset:        filter.Offset,
		Count:         len(subscriptions),
	})
}

// CreateSubscription godoc
// @Summary Create subscription
// @Description Creates a new subscription record.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body CreateSubscriptionRequest true "Subscription payload"
// @Success 201 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/subscriptions [post]
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
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to create subscription",
			"error", err,
			"user_id", params.UserID,
			"service_name", params.ServiceName,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create subscription",
		})
		return
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"subscription created",
		"id", subscription.ID,
		"user_id", subscription.UserID,
		"service_name", subscription.ServiceName,
		"price", subscription.Price)

	c.JSON(http.StatusCreated, toSubscriptionResponse(subscription))
}

// GetSubscriptionByID godoc
// @Summary Get subscription by id
// @Description Returns one subscription by id.
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription id"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/subscriptions/{id} [get]
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
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": subscriptionNotFound,
			})
			return
		}

		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to fetch subscription",
			"error", err,
			"id", id)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get subscription",
		})
		return
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"subscription found",
		"id", subscription.ID)

	c.JSON(http.StatusOK, toSubscriptionResponse(subscription))
}

// UpdateSubscription godoc
// @Summary Update subscription
// @Description Partially updates subscription by id. Only provided fields are changed.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription id"
// @Param request body UpdateSubscriptionRequest true "Subscription payload"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/subscriptions/{id} [put]
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

	currentSubscription, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": subscriptionNotFound,
			})
			return
		}

		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to fetch subscription before update",
			"error", err,
			"id", id)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch subscription",
		})
		return
	}

	params, err := buildUpdateSubscriptionParams(id, currentSubscription, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	subscription, err := h.repo.Update(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": subscriptionNotFound,
			})
			return
		}

		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to update subscription",
			"error", err,
			"id", id)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update subscription",
		})
		return
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"subscription updated",
		"id", subscription.ID,
		"user_id", subscription.UserID,
		"service_name", subscription.ServiceName,
		"price", subscription.Price)

	c.JSON(http.StatusOK, toSubscriptionResponse(subscription))
}

// DeleteSubscription godoc
// @Summary Delete subscription
// @Description Deletes subscription by id.
// @Tags subscriptions
// @Param id path int true "Subscription id"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": invalidIDError,
		})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": subscriptionNotFound,
			})
			return
		}

		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to delete subscription",
			"error", err,
			"id", id)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete subscription",
		})
		return
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"subscription deleted",
		"id", id)

	c.Status(http.StatusNoContent)
}

// GetTotalCost godoc
// @Summary Calculate total cost
// @Description Calculates total subscription cost for a selected period. Supports optional filters by user_id and service_name.
// @Tags subscriptions
// @Produce json
// @Param from query string true "Start period in MM-YYYY format"
// @Param to query string true "End period in MM-YYYY format"
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Subscription service name"
// @Success 200 {object} TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/subscriptions/total [get]
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
				"error": invalidUserIDError,
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
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to calculate total cost",
			"error", err,
			"from", from,
			"to", to,
			"user_id", userIDRaw,
			"service_name", serviceName,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to calculate total cost",
		})
		return
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"total cost calculated",
		"from", from,
		"to", to,
		"user_id", userIDRaw,
		"service_name", serviceName,
		"total", total)

	c.JSON(http.StatusOK, TotalCostResponse{
		From:        from,
		To:          to,
		UserID:      userIDRaw,
		ServiceName: serviceNameRaw,
		Total:       total,
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
		return model.CreateSubscriptionParams{}, errors.New(invalidUserIDError)
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		return model.CreateSubscriptionParams{}, errors.New("invalid start_date, expected format MM-YYYY")
	}

	endDateValue, hasEndDate, err := parseOptionalMonthYear(req.EndDate)
	if err != nil {
		return model.CreateSubscriptionParams{}, errors.New("invalid end_date, expected format MM-YYYY")
	}

	var endDate *time.Time
	if hasEndDate {
		endDate = &endDateValue
	}

	if endDate != nil && endDate.Before(startDate) {
		return model.CreateSubscriptionParams{}, errors.New("end date must be after than start date")
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
	current model.Subscription,
	req UpdateSubscriptionRequest,
) (model.UpdateSubscriptionParams, error) {
	if !hasUpdateFields(req) {
		return model.UpdateSubscriptionParams{}, errors.New("at least one field is required")
	}

	serviceName := current.ServiceName
	if req.ServiceName != nil {
		trimmedServiceName := strings.TrimSpace(*req.ServiceName)
		if trimmedServiceName == "" {
			return model.UpdateSubscriptionParams{}, errors.New("service_name is required")
		}

		serviceName = trimmedServiceName
	}

	price := current.Price
	if req.Price != nil {
		if *req.Price <= 0 {
			return model.UpdateSubscriptionParams{}, errors.New("price must be greater than zero")
		}

		price = *req.Price
	}

	userID := current.UserID
	if req.UserID != nil {
		parsedUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			return model.UpdateSubscriptionParams{}, errors.New(invalidUserIDError)
		}

		userID = parsedUserID
	}

	startDate := current.StartDate
	if req.StartDate != nil {
		parsedStartDate, err := parseMonthYear(*req.StartDate)
		if err != nil {
			return model.UpdateSubscriptionParams{}, errors.New("invalid start_date, expected format MM-YYYY")
		}

		startDate = parsedStartDate
	}

	endDate := current.EndDate
	if req.EndDate != nil {
		trimmedEndDate := strings.TrimSpace(*req.EndDate)
		if trimmedEndDate == "" {
			endDate = nil
		} else {
			parsedEndDate, err := parseMonthYear(trimmedEndDate)
			if err != nil {
				return model.UpdateSubscriptionParams{}, errors.New("invalid end_date, expected format MM-YYYY")
			}

			endDate = &parsedEndDate
		}
	}

	if endDate != nil && endDate.Before(startDate) {
		return model.UpdateSubscriptionParams{}, errors.New("end date must be after start date")
	}

	return model.UpdateSubscriptionParams{
		ID:          id,
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

func hasUpdateFields(req UpdateSubscriptionRequest) bool {
	return req.ServiceName != nil ||
		req.Price != nil ||
		req.UserID != nil ||
		req.StartDate != nil ||
		req.EndDate != nil
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

func parseOptionalMonthYear(value string) (time.Time, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false, nil
	}

	parsed, err := parseMonthYear(value)
	if err != nil {
		return time.Time{}, false, err
	}

	return parsed, true, nil
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

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}

func buildListSubscripionsFilter(c *gin.Context) (model.ListSubscriptionFilter, error) {
	limit, err := parseLimit(c.Query("limit"))
	if err != nil {
		return model.ListSubscriptionFilter{}, err
	}

	offset, err := parseOffset(c.Query("offset"))
	if err != nil {
		return model.ListSubscriptionFilter{}, err
	}

	var userID *uuid.UUID
	userIDRaw := strings.TrimSpace(c.Query("user_id"))
	if userIDRaw != "" {
		parsedUserID, err := uuid.Parse(userIDRaw)
		if err != nil {
			return model.ListSubscriptionFilter{}, errors.New(invalidUserIDError)
		}

		userID = &parsedUserID
	}

	var serviceName *string
	serviceNameRaw := strings.TrimSpace(c.Query("service_name"))
	if serviceNameRaw != "" {
		serviceName = &serviceNameRaw
	}

	return model.ListSubscriptionFilter{
		Limit:       limit,
		Offset:      offset,
		UserID:      userID,
		ServiceName: serviceName,
	}, nil
}

func parseLimit(rawLimit string) (int, error) {
	if rawLimit == "" {
		return defaultListLimit, nil
	}

	limit, err := strconv.Atoi(rawLimit)
	if err != nil {
		return 0, errors.New("limit must be a number")
	}

	if limit <= 0 {
		return 0, errors.New("limit must be greater than 0")
	}

	if limit > maxListLimit {
		return 0, errors.New("limit must be less than or equal to 100")
	}

	return limit, nil
}

func parseOffset(rawOffset string) (int, error) {
	if rawOffset == "" {
		return defaultOffset, nil
	}

	offset, err := strconv.Atoi(rawOffset)
	if err != nil {
		return 0, errors.New("offset must be a number")
	}

	if offset < 0 {
		return 0, errors.New("offset must be greater or equal to 0")
	}

	return offset, nil
}
