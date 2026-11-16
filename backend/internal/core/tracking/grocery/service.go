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

// ParseTextList parses pasted free text into lines matched against item history.
func ParseTextList(db *gorm.DB, text string) ([]ParsedLine, error) {
	lines := ParseList(text)
	if len(lines) == 0 {
		return []ParsedLine{}, nil
	}
	var known []struct {
		Name string
		Uses int
	}
	if err := db.Model(&GroceryItem{}).
		Select("name, COUNT(*) AS uses, MAX(created_at) AS last_used").
		Group("name").
		Order("uses DESC, last_used DESC").
		Scan(&known).Error; err != nil {
		return nil, err
	}
	knownItems := make([]KnownItem, 0, len(known))
	for _, k := range known {
		knownItems = append(knownItems, KnownItem{Name: k.Name, Count: k.Uses})
	}
	active, err := ListActiveItems(db)
	if err != nil {
		return nil, err
	}
	activeNames := make([]string, 0, len(active))
	for _, a := range active {
		activeNames = append(activeNames, a.Name)
	}
	return MatchLines(lines, knownItems, activeNames), nil
}

// CreateItems adds each name to the active list, skipping blanks and names
// already on the list (or repeated within the request). Returns created rows
// and the names that were skipped.
func CreateItems(db *gorm.DB, names []string) ([]GroceryItem, []string, error) {
	created := make([]GroceryItem, 0, len(names))
	skipped := make([]string, 0)
	err := db.Transaction(func(tx *gorm.DB) error {
		active, err := ListActiveItems(tx)
		if err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(active)+len(names))
		for _, a := range active {
			seen[strings.ToLower(a.Name)] = struct{}{}
		}
		for _, raw := range names {
			name := normalizeName(raw)
			if name == "" {
				continue
			}
			key := strings.ToLower(name)
			if _, exists := seen[key]; exists {
				skipped = append(skipped, name)
				continue
			}
			seen[key] = struct{}{}
			row := GroceryItem{Name: name}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			created = append(created, row)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return created, skipped, nil
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
