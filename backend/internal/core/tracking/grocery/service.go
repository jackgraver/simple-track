package grocery

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const suggestionLimit = 20

func normalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

func ListActiveItems(db *gorm.DB) ([]GroceryItem, error) {
	var rows []GroceryItem
	err := db.Where("completed_at IS NULL").Order("created_at ASC, id ASC").Find(&rows).Error
	return rows, err
}

func CreateItem(db *gorm.DB, name string) (*GroceryItem, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	row := GroceryItem{Name: name}
	if err := db.Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func CreateItemsOrExisting(db *gorm.DB, list []string) ([]string, error) {
	allItems, err := ListActiveItems(db);
	if err != nil {
		return nil, errors.New("Failed to load existing items")
	}

	// Create map of length of existing items list
	//   key = string 
	//   value = empty struct (no storage, when only care about key existence, Golang doesnt have sets so this is idiomatic way of doing it)
	allItemsMap := make(map[string]struct{}, len(allItems));

	// Populate map with keys for existing items to empty structs values
	for _, existingItem := range(allItems) {
		allItemsMap[strings.ToLower(existingItem.Name)] = struct{}{}
	} 

	var result []string
	// Check all requested items to see if they already exist 
	for _, newItem := range(list) {
		// go map lookup returns (value, exists), but discard value and do similar to inline err check to continue when new item doesnt exist 
		if _, exists := allItemsMap[strings.ToLower(newItem)]; !exists {
			result = append(result, newItem);
		}
	}

	return result, nil
}

func CompleteItem(db *gorm.DB, id uint) (*GroceryItem, error) {
	var row GroceryItem
	if err := db.First(&row, id).Error; err != nil {
		return nil, err
	}
	if row.CompletedAt == nil {
		now := time.Now()
		row.CompletedAt = &now
		if err := db.Save(&row).Error; err != nil {
			return nil, err
		}
	}
	return &row, nil
}

func DeleteItem(db *gorm.DB, id uint) error {
	var row GroceryItem
	if err := db.First(&row, id).Error; err != nil {
		return err
	}
	return db.Delete(&row).Error
}

func ListSuggestions(db *gorm.DB, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	rows := make([]struct {
		Name string
	}, 0)
	tx := db.Model(&GroceryItem{}).
		Select("name, MAX(completed_at) AS last_completed_at").
		Where("completed_at IS NOT NULL").
		Group("name").
		Order("last_completed_at DESC").
		Limit(suggestionLimit)
	if query != "" {
		tx = tx.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(query)+"%")
	}
	if err := tx.Scan(&rows).Error; err != nil {
		return nil, err
	}
	suggestions := make([]string, 0, len(rows))
	for _, row := range rows {
		suggestions = append(suggestions, row.Name)
	}
	return suggestions, nil
}
