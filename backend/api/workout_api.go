package api

import (
	"be-simpletracker/database/models"
	"be-simpletracker/database/services"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkoutFeature struct {
	BaseFeature[models.WorkoutModel]
}

func NewWorkoutFeature(db *gorm.DB) *WorkoutFeature {
    models.NewWorkoutModel(db)
    var feature = models.NewWorkoutModel(db)
    feature.MigrateDatabase()

	return &WorkoutFeature{
		BaseFeature[models.WorkoutModel]{
			db: db,
		},
	}
}

func (f *WorkoutFeature) SetEndpoints(router *gin.Engine) {
    group := router.Group("/workout")
    group.GET("/today", f.getWorkoutToday)
    group.GET("/month", f.getWorkoutMonth)
    group.GET("/all", f.getWorkoutAll)
    group.GET("/previous", f.getPreviousWorkout)
	group.POST("/exercise/log", f.logExercise)
    group.POST("/exercise/all-logged", f.checkAllLogged)
    group.GET("/exercises/all", f.getAllExercises)
    group.POST("/exercise/add", f.addExerciseToWorkout)
    group.DELETE("/exercise/remove", f.removeExerciseFromWorkout)
    group.GET("/exercise/progression/:id", f.getExerciseProgression)
    group.GET("/plans/all", f.getAllWorkoutPlans)
    group.POST("/plans/:id/exercises/add", f.addExerciseToPlan)
    group.DELETE("/plans/:id/exercises/remove", f.removeExerciseFromPlan)
    group.POST("/exercises/create", f.createExercise)
}

func (f *WorkoutFeature) getWorkoutToday(c *gin.Context) {
    day, err := services.GetToday(f.db, 0)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, day)
}


func (f *WorkoutFeature) getWorkoutMonth(c *gin.Context) {
    offsetStr := c.Query("monthoffset")
    offset, _ := strconv.Atoi(offsetStr)

    today := time.Now()
    target := today.AddDate(0, offset, 0)

    startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
    endOfMonth := startOfMonth.AddDate(0, 1, -1)

    start := startOfMonth.AddDate(0, 0, -int(startOfMonth.Weekday()))
    end := endOfMonth.AddDate(0, 0, 7-int(endOfMonth.Weekday()))

    // data, err := services.WorkoutRange(f.db, start, end)
    data, err := services.ObjectRange[*models.WorkoutLog](f.db, start, end)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "days":  data,
        "today": today,
        "range": gin.H{
            "start": start,
            "end":   end,
        },
        "month": target.Month(),
        "offset": offset,
    })
}

func (f *WorkoutFeature) getWorkoutAll(c *gin.Context) {
    days, err := services.GetAll(f.db)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, days)
}   

func totalVolume(log models.LoggedExercise) float32 {
	var total float32
	for _, s := range log.Sets {
		total += s.Weight * float32(s.Reps)
	}
	return total
}


//TODO: needs a lot more work, I think we need more data before were able to do more complex stuff like this
/*
		prevLog, err := services.GetPreviousExerciseLog(f.db, today.Date, exercise.Name, -1)
		if err != nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
			return
		}

		currentVol := totalVolume(exerciseLog)
		prevVol := totalVolume(prevLog)

		if prevVol > 0 {
			difference := ((currentVol - prevVol) / prevVol) * 100
			if math.Abs(float64(difference)) >= 5 {
				exerciseLog.PercentChange = float32(difference)
			}
		}
*/
type ExerciseGroup struct {
    Planned         *models.Exercise       `json:"planned,omitempty"`
    Logged          *models.LoggedExercise `json:"logged,omitempty"`
    Previous        *models.LoggedExercise `json:"previous,omitempty"`
}
func (f *WorkoutFeature) getPreviousWorkout(c *gin.Context) {
    offsetStr := c.Query("offset")
    offset, _ := strconv.Atoi(offsetStr)

    today, err := services.GetToday(f.db, offset)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }

    logged := today.Exercises
    planned := today.WorkoutPlan.Exercises

    // Build a map for logged exercises keyed by name
    loggedMap := make(map[string]models.LoggedExercise)
    for _, l := range logged {
        if l.Exercise != nil {
            loggedMap[l.Exercise.Name] = l
        }
    }

    // Prepare grouped list
    results := make([]ExerciseGroup, 0)

    // First, handle planned exercises (ensures order by plan)
    for _, p := range planned {
        group := ExerciseGroup{Planned: &p}

        // if already logged, attach it
        if log, ok := loggedMap[p.Name]; ok {
            group.Logged = &log
            delete(loggedMap, p.Name) // remove to avoid duplicates
        }

        // get previous log
        prev, err := services.GetPreviousExerciseLog(f.db, today.Date, p.Name, 0)
        if err == nil {
            group.Previous = &prev
        }

        results = append(results, group)
    }

    // Any leftover logged exercises not part of the plan (edge case)
    for _, l := range loggedMap {
        prev, err := services.GetPreviousExerciseLog(f.db, today.Date, l.Exercise.Name, 0)
        if err == nil {
            results = append(results, ExerciseGroup{
                Logged:   &l,
                Previous: &prev,
            })
        } else {
            results = append(results, ExerciseGroup{Logged: &l})
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "day":        today,
        "previous_exercises":  results,
    })
}


type LogExerciseRequest struct {
	Log models.LoggedExercise `json:"exercise"`
    Type string `json:"type"`
}

func (f *WorkoutFeature) logExercise(c *gin.Context) {
	var request LogExerciseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
    }

    switch request.Type {
        case "previous":
            request.Log.ID = 0
            for i := range request.Log.Sets {
                request.Log.Sets[i].LoggedExerciseID = 0
                request.Log.Sets[i].ID = 0
            }
            err := services.LogExercise(f.db, &request.Log)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
        case "logged":
            err := services.UpdateLoggedExercise(f.db, request.Log)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
    }

	c.JSON(http.StatusOK, gin.H{"exercise": request.Log})
}

func (f *WorkoutFeature) checkAllLogged(c *gin.Context) {
    today, err := services.GetToday(f.db, 0)
    if err != nil {
        c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
        return
    }

    if len(today.Exercises) == len(today.WorkoutPlan.Exercises) {
        fmt.Println("All logged!")
        c.JSON(http.StatusOK, gin.H{"all_logged": true})
        return
    }

    fmt.Println("Not all logged!")
    c.JSON(http.StatusOK, gin.H{"all_logged": false})
}

func (f *WorkoutFeature) getAllExercises(c *gin.Context) {
    excludeParam := c.QueryArray("exclude")
    var excludeIDs []uint
    for _, idStr := range excludeParam {
        if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
            excludeIDs = append(excludeIDs, uint(id))
        }
    }

    exercises, err := services.GetAllExercises(f.db, excludeIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"exercises": exercises})
}

type AddExerciseRequest struct {
    ExerciseID uint `json:"exercise_id"`
}

func (f *WorkoutFeature) addExerciseToWorkout(c *gin.Context) {
    var request AddExerciseRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    today, err := services.GetToday(f.db, 0)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Check if exercise already exists in workout
    for _, ex := range today.Exercises {
        if ex.ExerciseID == request.ExerciseID {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise already in workout"})
            return
        }
    }

    // Create a new logged exercise entry (empty, ready to be logged)
    newExercise := models.LoggedExercise{
        WorkoutLogID: today.ID,
        ExerciseID:   request.ExerciseID,
        Sets:         []models.LoggedSet{},
        Notes:        "",
    }

    err = services.LogExercise(f.db, &newExercise)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Reload the exercise with Exercise relation using the ID that was set by Create
    var createdExercise models.LoggedExercise
    err = f.db.Preload("Exercise").Preload("Sets").Where("id = ?", newExercise.ID).First(&createdExercise).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"exercise": createdExercise})
}

type RemoveExerciseRequest struct {
    ExerciseID uint `json:"exercise_id"`
}

func (f *WorkoutFeature) removeExerciseFromWorkout(c *gin.Context) {
    var request RemoveExerciseRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    today, err := services.GetToday(f.db, 0)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Find and delete the logged exercise
    var loggedExercise models.LoggedExercise
    err = f.db.Where("workout_log_id = ? AND exercise_id = ?", today.ID, request.ExerciseID).First(&loggedExercise).Error
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found in workout"})
        return
    }

    // Delete the logged exercise (sets will be cascade deleted)
    err = f.db.Delete(&loggedExercise).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

func (f *WorkoutFeature) getExerciseProgression(c *gin.Context) {
    exerciseIDStr := c.Param("id")
    exerciseID, err := strconv.ParseUint(exerciseIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID"})
        return
    }

    progression, err := services.GetExerciseProgression(f.db, uint(exerciseID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"progression": progression})
}

func (f *WorkoutFeature) getAllWorkoutPlans(c *gin.Context) {
    plans, err := services.GetAllWorkoutPlans(f.db)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"plans": plans})
}

type AddExerciseToPlanRequest struct {
    ExerciseID uint `json:"exercise_id"`
}

func (f *WorkoutFeature) addExerciseToPlan(c *gin.Context) {
    planIDStr := c.Param("id")
    planID, err := strconv.ParseUint(planIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
        return
    }

    var request AddExerciseToPlanRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    err = services.AddExerciseToPlan(f.db, uint(planID), request.ExerciseID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

type RemoveExerciseFromPlanRequest struct {
    ExerciseID uint `json:"exercise_id"`
}

func (f *WorkoutFeature) removeExerciseFromPlan(c *gin.Context) {
    planIDStr := c.Param("id")
    planID, err := strconv.ParseUint(planIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
        return
    }

    var request RemoveExerciseFromPlanRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    err = services.RemoveExerciseFromPlan(f.db, uint(planID), request.ExerciseID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}

type CreateExerciseRequest struct {
    Name        string `json:"name"`
    RepRollover uint   `json:"rep_rollover"`
}

func (f *WorkoutFeature) createExercise(c *gin.Context) {
    var request CreateExerciseRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if request.Name == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise name is required"})
        return
    }

    if request.RepRollover == 0 {
        request.RepRollover = 10
    }

    exercise, err := services.CreateExercise(f.db, request.Name, request.RepRollover)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}