package controller

import (
	finance_service "be-simpletracker/internal/core/finance/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAllAccounts(c *gin.Context) {
	rows, err := finance_service.ListAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": rows})
}

type createAccountBody struct {
	Name    string  `json:"name"`
	Balance float32 `json:"balance"`
}

func CreateAccount(c *gin.Context) {
	var body createAccountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := finance_service.CreateAccount(body.Name, body.Balance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": a})
}

func GetAllTransactions(c *gin.Context) {
	rows, err := finance_service.ListTransactions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": rows})
}

func GetAllCategories(c *gin.Context) {
	rows, err := finance_service.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": rows})
}

type createTransactionBody struct {
	AccountID  uint    `json:"account_id"`
	Amount     float32 `json:"amount"`
	Date       string  `json:"date"`
	CategoryID uint    `json:"category_id"`
}

func CreateTransaction(c *gin.Context) {
	var body createTransactionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(body.AccountID, body.Amount, body.Date, body.CategoryID)
	if body.CategoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_id is required"})
		return
	}
	if body.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}
	var at time.Time
	parsed, err := time.Parse(time.RFC3339, body.Date)
	if err != nil {
		var err2 error
		parsed, err2 = time.ParseInLocation("2006-01-02", body.Date, time.Local)
		if err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date: use RFC3339 or YYYY-MM-DD"})
			return
		}
	}
	at = parsed
	t, err := finance_service.CreateTransaction(body.AccountID, body.Amount, at, body.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction": t})
}
