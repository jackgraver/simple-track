package repository_test

import (
	"be-simpletracker/database/models"
	"be-simpletracker/database/repository"
	"context"
	"time"
)

// This file demonstrates how to use the generic repository.
// It's not meant to be run as tests, but as documentation.

func ExampleBasicUsage() {
	// Assume db is a *gorm.DB instance
	var db interface{} // placeholder
	_ = db

	// Create a repository for WorkoutLog
	// repo := repository.NewGormRepository[*models.WorkoutLog](db.(*gorm.DB))

	// Get by ID with default preloads
	// log, err := repo.GetByID(context.Background(), 1)

	// Get by ID with specific preloads
	// log, err := repo.GetByID(context.Background(), 1,
	//     repository.WithPreloads("Exercises", "Cardio"),
	// )

	// Get by ID with no preloads (faster for simple queries)
	// log, err := repo.GetByID(context.Background(), 1,
	//     repository.WithNoPreloads(),
	// )
}

func ExampleGetAll() {
	// Get all workout logs ordered by date descending
	// logs, err := repo.GetAll(context.Background(),
	//     repository.WithOrderByDesc("date"),
	// )

	// Get all excluding specific IDs
	// logs, err := repo.GetAll(context.Background(),
	//     repository.WithExcludeIDs(1, 2, 3),
	// )

	// Get all with a filter
	// logs, err := repo.GetAll(context.Background(),
	//     repository.WithFilter("workout_plan_id", 5),
	// )
}

func ExamplePagination() {
	// Get page 2 with 10 items per page
	// result, err := repo.GetAllPaginated(context.Background(), 2, 10)
	//
	// result.Data       // []WorkoutLog - the items
	// result.Total      // Total count
	// result.Page       // Current page (2)
	// result.PageSize   // Items per page (10)
	// result.TotalPages // Total pages
	// result.HasNext    // bool
	// result.HasPrev    // bool
}

func ExampleDateRangeQueries() {
	// Get logs for the past week
	// now := time.Now()
	// weekAgo := now.AddDate(0, 0, -7)
	// logs, err := repo.GetByDateRange(context.Background(), weekAgo, now)

	// Get logs for a specific month with pagination
	// start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)
	// end := time.Date(2025, time.January, 31, 23, 59, 59, 0, time.Local)
	// result, err := repo.GetByDateRangePaginated(context.Background(), start, end, 1, 20)
}

func ExampleCombinedOptions() {
	// Combine multiple options
	// logs, err := repo.GetAll(context.Background(),
	//     repository.WithPreloads("Exercises.Sets", "Cardio"),
	//     repository.WithOrderByDesc("date"),
	//     repository.WithFilter("workout_plan_id", 5),
	//     repository.WithLimit(10),
	// )
}

func ExampleTransactions() {
	// Execute operations in a transaction
	// err := repo.Transaction(context.Background(), func(txRepo repository.Repository[*models.WorkoutLog]) error {
	//     // All operations use the transaction
	//     log, err := txRepo.GetByID(context.Background(), 1)
	//     if err != nil {
	//         return err // Transaction rolls back
	//     }
	//
	//     log.Date = time.Now()
	//     return txRepo.Update(context.Background(), log)
	// })
}

func ExampleCustomRepository() {
	// For domain-specific queries, embed GormRepository in your own type:
	//
	// type WorkoutLogRepository struct {
	//     *repository.GormRepository[*models.WorkoutLog]
	// }
	//
	// func NewWorkoutLogRepository(db *gorm.DB) *WorkoutLogRepository {
	//     return &WorkoutLogRepository{
	//         GormRepository: repository.NewGormRepository[*models.WorkoutLog](db),
	//     }
	// }
	//
	// // Add domain-specific methods
	// func (r *WorkoutLogRepository) GetTodaysWorkout(ctx context.Context) (*models.WorkoutLog, error) {
	//     today := time.Now().Truncate(24 * time.Hour)
	//     return r.GetByDate(ctx, today)
	// }
	//
	// func (r *WorkoutLogRepository) GetPreviousWorkoutByPlan(ctx context.Context, planID uint, beforeDate time.Time) (*models.WorkoutLog, error) {
	//     return r.FindOne(ctx,
	//         repository.WithFilter("workout_plan_id", planID),
	//         repository.WithDateUntil(beforeDate),
	//         repository.WithOrderByDesc("date"),
	//     )
	// }
}

// Demonstrates the pattern for using repositories in services
func ExampleServicePattern() {
	// type WorkoutService struct {
	//     repo repository.Repository[*models.WorkoutLog]
	// }
	//
	// func NewWorkoutService(repo repository.Repository[*models.WorkoutLog]) *WorkoutService {
	//     return &WorkoutService{repo: repo}
	// }
	//
	// func (s *WorkoutService) GetTodayWorkout(ctx context.Context, offset int) (*models.WorkoutLog, error) {
	//     date := utils.ZerodTime(offset)
	//     return s.repo.(*repository.GormRepository[*models.WorkoutLog]).GetByDate(ctx, date)
	// }
}

// Prevent unused import errors
var (
	_ = context.Background
	_ = time.Now
	_ = repository.NewGormRepository[*models.WorkoutLog]
	_ = models.WorkoutLog{}
)
