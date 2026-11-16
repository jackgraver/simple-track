package service

import (
	"sort"
	"time"

	"be-simpletracker/internal/core/money/models"

	"gorm.io/gorm"
)

type NetWorthPoint struct {
	Date        time.Time `json:"date"`
	Assets      float64   `json:"assets"`
	Liabilities float64   `json:"liabilities"`
	NetWorth    float64   `json:"net_worth"`
}

// NetWorthHistory returns net worth on every day any balance was recorded.
// Each account contributes its most recent balance on or before that day,
// and stops counting once it is archived.
func NetWorthHistory(db *gorm.DB) ([]NetWorthPoint, error) {
	var accounts []models.Account
	if err := db.Find(&accounts).Error; err != nil {
		return nil, err
	}
	var balances []models.AccountBalance
	if err := db.Order("date ASC").Find(&balances).Error; err != nil {
		return nil, err
	}
	return buildNetWorthHistory(accounts, balances), nil
}

func buildNetWorthHistory(accounts []models.Account, balances []models.AccountBalance) []NetWorthPoint {
	byID := make(map[uint]models.Account, len(accounts))
	for _, account := range accounts {
		byID[account.ID] = account
	}
	sort.SliceStable(balances, func(i, j int) bool { return balances[i].Date.Before(balances[j].Date) })

	points := []NetWorthPoint{}
	current := map[uint]float64{}
	for i := 0; i < len(balances); {
		day := balances[i].Date
		for ; i < len(balances) && balances[i].Date.Equal(day); i++ {
			if _, ok := byID[balances[i].AccountID]; ok {
				current[balances[i].AccountID] = balances[i].Balance
			}
		}
		point := NetWorthPoint{Date: day}
		for accountID, balance := range current {
			account := byID[accountID]
			if account.ArchivedAt != nil && account.ArchivedAt.Before(day) {
				continue
			}
			if account.Kind.IsLiability() {
				point.Liabilities += balance
			} else {
				point.Assets += balance
			}
		}
		point.Assets = roundCents(point.Assets)
		point.Liabilities = roundCents(point.Liabilities)
		point.NetWorth = roundCents(point.Assets - point.Liabilities)
		points = append(points, point)
	}
	return points
}
