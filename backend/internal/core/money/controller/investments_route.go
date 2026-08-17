package controller

import (
	"errors"
	"net/http"
	"strconv"

	authmodels "be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/core/money/service"
	"be-simpletracker/internal/core/tracking/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InvestmentHandler struct {
	db *gorm.DB
}

func RegisterInvestmentRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := InvestmentHandler{db: db}
	group.GET("", h.listAccounts)
	group.POST("", h.createAccount)
	group.PATCH("/:id", h.updateAccount)
	group.DELETE("/:id", h.deleteAccount)
	group.GET("/:id/deposits", h.listDeposits)
	group.POST("/:id/deposits", h.createDeposit)
	group.DELETE("/:id/deposits/:deposit_id", h.deleteDeposit)
	accountTypes := group.Group("/account-types")
	accountTypes.GET("", h.listAccountTypes)
	accountTypes.POST("", h.createAccountType)
	accountTypes.PATCH("/:id", h.updateAccountType)
	accountTypes.DELETE("/:id", h.deleteAccountType)
	accountTypes.PUT("/:id/contribution-rules/:year", h.upsertContributionRule)
	accountTypes.DELETE("/:id/contribution-rules/:year", h.deleteContributionRule)
}

type accountBody struct {
	Name                    *string  `json:"name"`
	InvestmentAccountTypeID *uint    `json:"investment_account_type_id"`
	CurrentBalance          *float64 `json:"current_balance"`
}

func (h *InvestmentHandler) listAccounts(c *gin.Context) {
	birthYear, ok := h.currentUserBirthYear(c)
	if !ok {
		return
	}
	accounts, err := service.ListInvestmentAccounts(h.db, birthYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

func (h *InvestmentHandler) createAccount(c *gin.Context) {
	var body accountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Name == nil || body.CurrentBalance == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and current_balance are required"})
		return
	}
	account, err := service.CreateInvestmentAccount(h.db, *body.Name, body.InvestmentAccountTypeID, *body.CurrentBalance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"account": account})
}

func (h *InvestmentHandler) updateAccount(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body accountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	account, err := service.UpdateInvestmentAccount(h.db, id, body.Name, body.InvestmentAccountTypeID, body.CurrentBalance)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": account})
}

func (h *InvestmentHandler) deleteAccount(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteInvestmentAccount(h.db, id); err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type accountTypeBody struct {
	Name                 string `json:"name"`
	ContributionStartAge *int   `json:"contribution_start_age"`
}

func (h *InvestmentHandler) listAccountTypes(c *gin.Context) {
	birthYear, ok := h.currentUserBirthYear(c)
	if !ok {
		return
	}
	accountTypes, err := service.ListInvestmentAccountTypes(h.db, birthYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"account_types": accountTypes})
}

func (h *InvestmentHandler) createAccountType(c *gin.Context) {
	var body accountTypeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accountType, err := service.CreateInvestmentAccountType(h.db, body.Name, body.ContributionStartAge)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"account_type": accountType})
}

func (h *InvestmentHandler) updateAccountType(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body accountTypeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accountType, err := service.UpdateInvestmentAccountType(h.db, id, body.Name, body.ContributionStartAge)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"account_type": accountType})
}

func (h *InvestmentHandler) deleteAccountType(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteInvestmentAccountType(h.db, id); err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type depositBody struct {
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
}

func (h *InvestmentHandler) listDeposits(c *gin.Context) {
	accountID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	deposits, err := service.ListInvestmentDeposits(h.db, accountID)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deposits": deposits})
}

func (h *InvestmentHandler) createDeposit(c *gin.Context) {
	accountID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body depositBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := common.ParseDateString(body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	deposit, err := service.CreateInvestmentDeposit(h.db, accountID, body.Amount, date)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"deposit": deposit})
}

func (h *InvestmentHandler) deleteDeposit(c *gin.Context) {
	accountID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	depositID, ok := parseUintParam(c, "deposit_id")
	if !ok {
		return
	}
	if err := service.DeleteInvestmentDeposit(h.db, accountID, depositID); err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type contributionRuleBody struct {
	AnnualLimit *float64 `json:"annual_limit"`
}

func (h *InvestmentHandler) upsertContributionRule(c *gin.Context) {
	accountTypeID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	year, ok := parseYearParam(c)
	if !ok {
		return
	}
	var body contributionRuleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.AnnualLimit == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "annual_limit is required"})
		return
	}
	rule, err := service.UpsertContributionRule(h.db, accountTypeID, year, *body.AnnualLimit)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"contribution_rule": rule})
}

func (h *InvestmentHandler) deleteContributionRule(c *gin.Context) {
	accountTypeID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	year, ok := parseYearParam(c)
	if !ok {
		return
	}
	if err := service.DeleteContributionRule(h.db, accountTypeID, year); err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return uint(value), true
}

func parseYearParam(c *gin.Context) (int, bool) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil || year < 1900 || year > 9999 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return 0, false
	}
	return year, true
}

func respondAccountError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "account or record not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func (h *InvestmentHandler) currentUserBirthYear(c *gin.Context) (*int, bool) {
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return nil, false
	}
	var user authmodels.User
	if err := h.db.Select("birth_year").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, true
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil, false
	}
	return user.BirthYear, true
}
