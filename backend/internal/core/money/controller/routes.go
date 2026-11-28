package controller

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/database"
	"errors"
	"net/http"
	"strconv"
	"time"

	authmodels "be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/core/money/models"
	"be-simpletracker/internal/core/money/service"
	"be-simpletracker/internal/core/tracking/common"
	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func RegisterRoutes(group *gin.RouterGroup) {
	h := Handler{}

	accounts := group.Group("/accounts")
	accounts.GET("", h.listAccounts)
	accounts.POST("", h.createAccount)
	accounts.GET("/:id", h.getAccount)
	accounts.PUT("/:id", h.updateAccount)
	accounts.DELETE("/:id", h.deleteAccount)
	accounts.GET("/:id/balances", h.listBalances)
	accounts.PUT("/:id/balances", h.setBalance)
	accounts.DELETE("/:id/balances/:balance_id", h.deleteBalance)
	accounts.GET("/:id/transactions", h.listTransactions)
	accounts.POST("/:id/transactions", h.createTransaction)
	accounts.DELETE("/:id/transactions/:transaction_id", h.deleteTransaction)

	group.PUT("/balances", h.setBalances)
	group.GET("/net-worth/history", h.netWorthHistory)

	types := group.Group("/contribution-types")
	types.GET("", h.listContributionTypes)
	types.POST("", h.createContributionType)
	types.GET("/presets", h.listPresets)
	types.POST("/presets", h.createFromPreset)
	types.GET("/:id", h.getContributionType)
	types.PUT("/:id", h.updateContributionType)
	types.PUT("/:id/room", h.setContributionRoom)
	types.DELETE("/:id", h.deleteContributionType)
	types.PUT("/:id/limits/:year", h.upsertLimit)
	types.DELETE("/:id/limits/:year", h.deleteLimit)
}

// ---- Accounts ----

func (h *Handler) listAccounts(c *gin.Context) {
	accounts, err := service.ListAccounts(c.Request.Context(), c.Query("archived") == "true")
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts, "kinds": models.AccountKinds})
}

func (h *Handler) getAccount(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	account, err := service.GetAccount(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": account})
}

type createAccountBody struct {
	service.AccountInput
	OpeningBalance *float64 `json:"opening_balance"`
	BalanceDate    string   `json:"balance_date"`
}

func (h *Handler) createAccount(c *gin.Context) {
	var body createAccountBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	date, ok := parseDate(c, body.BalanceDate)
	if !ok {
		return
	}
	account, err := service.CreateAccount(c.Request.Context(), body.AccountInput, body.OpeningBalance, date)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"account": account})
}

func (h *Handler) updateAccount(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body service.AccountInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	account, err := service.UpdateAccount(c.Request.Context(), id, body)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"account": account})
}

func (h *Handler) deleteAccount(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteAccount(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Balances ----

func (h *Handler) listBalances(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	balances, err := service.ListBalances(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"balances": balances})
}

type balanceBody struct {
	Date    string   `json:"date"`
	Balance *float64 `json:"balance"`
}

func (h *Handler) setBalance(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body balanceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if body.Balance == nil {
		apierr.BadRequest(c, "balance is required")
		return
	}
	date, ok := parseDate(c, body.Date)
	if !ok {
		return
	}
	if err := service.SetBalance(c.Request.Context(), id, date, *body.Balance); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type balancesBody struct {
	Date     string                 `json:"date"`
	Balances []service.BalanceEntry `json:"balances"`
}

func (h *Handler) setBalances(c *gin.Context) {
	var body balancesBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	date, ok := parseDate(c, body.Date)
	if !ok {
		return
	}
	if err := service.SetBalances(c.Request.Context(), date, body.Balances); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) deleteBalance(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	balanceID, ok := parseUintParam(c, "balance_id")
	if !ok {
		return
	}
	if err := service.DeleteBalance(c.Request.Context(), id, balanceID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) netWorthHistory(c *gin.Context) {
	points, err := service.NetWorthHistory(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

// ---- Transactions ----

func (h *Handler) listTransactions(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	transactions, err := service.ListTransactions(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

type transactionBody struct {
	service.TransactionInput
	Date string `json:"date"`
}

func (h *Handler) createTransaction(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body transactionBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	date, ok := parseDate(c, body.Date)
	if !ok {
		return
	}
	input := body.TransactionInput
	input.Date = date
	transaction, err := service.CreateTransaction(c.Request.Context(), id, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"transaction": transaction})
}

func (h *Handler) deleteTransaction(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	transactionID, ok := parseUintParam(c, "transaction_id")
	if !ok {
		return
	}
	if err := service.DeleteTransaction(c.Request.Context(), id, transactionID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Contribution types ----

func (h *Handler) listContributionTypes(c *gin.Context) {
	birthYear, ok := h.currentUserBirthYear(c)
	if !ok {
		return
	}
	types, err := service.ListContributionTypes(c.Request.Context(), birthYear, clock.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"contribution_types": types, "birth_year": birthYear})
}

func (h *Handler) getContributionType(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	birthYear, ok := h.currentUserBirthYear(c)
	if !ok {
		return
	}
	contributionType, err := service.GetContributionType(c.Request.Context(), id, birthYear, clock.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"contribution_type": contributionType, "birth_year": birthYear})
}

func (h *Handler) listPresets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"presets": service.ListContributionPresets()})
}

func (h *Handler) createFromPreset(c *gin.Context) {
	var body service.PresetInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	contributionType, err := service.CreateContributionTypeFromPreset(c.Request.Context(), body)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"contribution_type": contributionType})
}

func (h *Handler) createContributionType(c *gin.Context) {
	var body service.ContributionTypeCreateInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	contributionType, err := service.CreateContributionType(c.Request.Context(), body)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"contribution_type": contributionType})
}

func (h *Handler) updateContributionType(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body service.ContributionTypeInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	contributionType, err := service.UpdateContributionType(c.Request.Context(), id, body)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"contribution_type": contributionType})
}

// setContributionRoom saves the caller's own room figures (start year, opening room) for a type.
// They are per user, so anyone signed in may set theirs on a shared type.
func (h *Handler) setContributionRoom(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var body service.RoomInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if err := service.SetContributionRoom(c.Request.Context(), id, body); err != nil {
		respondError(c, err)
		return
	}
	birthYear, ok := h.currentUserBirthYear(c)
	if !ok {
		return
	}
	contributionType, err := service.GetContributionType(c.Request.Context(), id, birthYear, clock.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"contribution_type": contributionType})
}

func (h *Handler) deleteContributionType(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteContributionType(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type limitBody struct {
	AnnualLimit *float64 `json:"annual_limit"`
}

func (h *Handler) upsertLimit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	year, ok := parseYearParam(c)
	if !ok {
		return
	}
	var body limitBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if body.AnnualLimit == nil {
		apierr.BadRequest(c, "annual_limit is required")
		return
	}
	limit, err := service.UpsertContributionLimit(c.Request.Context(), id, year, *body.AnnualLimit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"limit": limit})
}

func (h *Handler) deleteLimit(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	year, ok := parseYearParam(c)
	if !ok {
		return
	}
	if err := service.DeleteContributionLimit(c.Request.Context(), id, year); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---- Helpers ----

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return uint(value), true
}

func parseYearParam(c *gin.Context) (int, bool) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil || year < 1900 || year > 9999 {
		apierr.BadRequest(c, "invalid year")
		return 0, false
	}
	return year, true
}

// parseDate parses YYYY-MM-DD; an empty string means today.
func parseDate(c *gin.Context, value string) (time.Time, bool) {
	date, err := common.ParseDateString(value)
	if err != nil {
		apierr.BadRequest(c, "invalid date, expected YYYY-MM-DD")
		return time.Time{}, false
	}
	return date, true
}

// respondError maps a service error to a response: 400 for bad input, 403 for a shared row the
// caller can see but may not change, 404 for a missing or someone else's row, else a logged 500.
func respondError(c *gin.Context, err error) {
	var validation service.ValidationError
	switch {
	case errors.As(err, &validation):
		apierr.BadRequest(c, validation.Message)
	case errors.Is(err, service.ErrForbidden):
		apierr.Forbidden(c, "not allowed")
	default:
		apierr.NotFoundOrInternal(c, err)
	}
}

func (h *Handler) currentUserBirthYear(c *gin.Context) (*int, bool) {
	id, ok := database.CurrentIdentity(c.Request.Context())
	if !ok {
		apierr.Unauthorized(c, "Not authenticated")
		return nil, false
	}
	var user authmodels.User
	if err := database.From(c.Request.Context()).Select("birth_year").First(&user, id.UserID).Error; err != nil {
		if database.IsNotFound(err) {
			return nil, true
		}
		apierr.Internal(c, err)
		return nil, false
	}
	return user.BirthYear, true
}
