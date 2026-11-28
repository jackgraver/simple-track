package grocery

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/database"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const suggestionLimit = 20

// ErrNameRequired is returned when an item name is empty after trimming.
var ErrNameRequired = errors.New("name is required")

func normalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

// Every query below is the current user's (database.Owned), and creates are stamped with them.

func ListActiveItems(ctx context.Context) ([]GroceryItem, error) {
	return listActiveItems(ctx, database.From(ctx))
}

func listActiveItems(ctx context.Context, db *gorm.DB) ([]GroceryItem, error) {
	var rows []GroceryItem
	err := db.Scopes(database.Owned(ctx)).Where("completed_at IS NULL").Order("created_at ASC, id ASC").Find(&rows).Error
	return rows, err
}

// CreateItem adds an item. A non-empty clientID makes the call idempotent: if the user already has
// an item with that id (even completed or deleted since), it is returned and nothing is created.
// Another user's item with the same client_id is not a match.
func CreateItem(ctx context.Context, name, clientID string) (*GroceryItem, error) {
	db := database.From(ctx)
	name = normalizeName(name)
	if name == "" {
		return nil, ErrNameRequired
	}
	row := GroceryItem{Name: name}
	if clientID == "" {
		if err := db.Omit(clause.Associations).Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}
	find := func() (*GroceryItem, error) {
		var existing GroceryItem
		err := db.Unscoped().Scopes(database.Owned(ctx)).Where("client_id = ?", clientID).First(&existing).Error
		if err != nil {
			return nil, err
		}
		return &existing, nil
	}
	if existing, err := find(); err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	row.ClientID = &clientID
	if err := database.CreateInSavepoint(db, &row); err != nil {
		// Lost a race with a concurrent retry of the same request.
		if existing, ferr := find(); ferr == nil {
			return existing, nil
		}
		return nil, err
	}
	return &row, nil
}

// ParseTextList parses pasted free text into lines matched against item history.
func ParseTextList(ctx context.Context, text string) ([]ParsedLine, error) {
	db := database.From(ctx)
	lines := ParseList(text)
	if len(lines) == 0 {
		return []ParsedLine{}, nil
	}
	var known []struct {
		Name string
		Uses int
	}
	if err := db.Model(&GroceryItem{}).Scopes(database.Owned(ctx)).
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
	active, err := ListActiveItems(ctx)
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
func CreateItems(ctx context.Context, names []string) ([]GroceryItem, []string, error) {
	db := database.From(ctx)
	created := make([]GroceryItem, 0, len(names))
	skipped := make([]string, 0)
	err := db.Transaction(func(tx *gorm.DB) error {
		active, err := listActiveItems(ctx, tx)
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
			if err := tx.Omit(clause.Associations).Create(&row).Error; err != nil {
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

// findItem loads the current user's item by id or client_id (ErrNotFound for a missing item or
// someone else's).
func findItem(ctx context.Context, db *gorm.DB, ref itemRef) (GroceryItem, error) {
	var row GroceryItem
	db = db.Scopes(database.Owned(ctx))
	if ref.clientID != "" {
		return row, db.Where("client_id = ?", ref.clientID).First(&row).Error
	}
	return row, db.First(&row, ref.id).Error
}

// CompleteItem marks the item done. Idempotent: completing a completed item returns it unchanged.
func CompleteItem(ctx context.Context, ref itemRef) (*GroceryItem, error) {
	db := database.From(ctx)
	row, err := findItem(ctx, db, ref)
	if err != nil {
		return nil, err
	}
	if row.CompletedAt == nil {
		now := clock.Now()
		row.CompletedAt = &now
		if err := db.Omit(clause.Associations).Save(&row).Error; err != nil {
			return nil, err
		}
	}
	return &row, nil
}

func DeleteItem(ctx context.Context, ref itemRef) error {
	db := database.From(ctx)
	row, err := findItem(ctx, db, ref)
	if err != nil {
		return err
	}
	return db.Scopes(database.Owned(ctx)).Delete(&row).Error
}

func ListSuggestions(ctx context.Context, query string) ([]string, error) {
	db := database.From(ctx)
	query = strings.TrimSpace(query)
	rows := make([]struct {
		Name string
	}, 0)
	tx := db.Model(&GroceryItem{}).Scopes(database.Owned(ctx)).
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
