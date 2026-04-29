package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Subscription struct {
	ID          int       `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   time.Time `json:"start_date"`
	// pointer can be empty
	EndDate *time.Time `json:"end_date,omitempty"`
}

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

func ListSubscriptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"subscriptions": []Subscription{},
	})
}

func CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date, format should be MM-YYYY"})
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		parsedEndDate, err := parseMonthYear(req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date, format should be MM-YYYY"})
			return
		}
		endDate = &parsedEndDate
	}

	if req.ServiceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_name is required"})
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price might be greater than 0"})
	}

	subscription := Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	c.JSON(http.StatusCreated, subscription)
}

func GetSubscriptionByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "get subscription by id",
	})
}

func UpdateSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "update subscription by id",
	})
}

func DeleteSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "delete subscription by id",
	})
}

func GetTotalCost(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	userID := c.Query("user_id")
	serviceName := c.Query("service_name")

	fromDate, err := parseMonthYear(from)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date, format should be MM-YYYY"})
		return
	}

	toDate, err := parseMonthYear(to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date, format should be MM-YYYY"})
		return
	}

	if fromDate.After(toDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end date is before start date"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from":         from,
		"to":           to,
		"user":         userID,
		"service_name": serviceName,
		"total_cost":   0,
	})
}

func parseID(rawID string) (int, error) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		return 0, err
	}

	if id < 0 {
		return 0, errors.New("id might be greater than 0")
	}

	return id, nil
}

func parseMonthYear(value string) (time.Time, error) {
	return time.Parse("01-2006", value)
}
