package workout

import (
	"be-simpletracker/internal/core/workout/controller"
	"be-simpletracker/internal/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	dayOffsetMiddleware := utils.DayOffsetMiddleware()

	group := router.Group("/workout", authMiddleware)
	{
		programs := group.Group("/programs")
		{
			programs.GET("", controller.GetAllWorkoutPrograms)
			programs.POST("", controller.CreateWorkoutProgram)
			programs.PATCH("/:id", controller.RenameWorkoutProgram)
			programs.POST("/:id/activate", controller.ActivateWorkoutProgram)
			programs.POST("/:id/plans", controller.CreateWorkoutPlan)
		}
		plans := group.Group("/plans")
		{
			plans.GET("/all", controller.GetAllWorkoutPlans)
			plans.POST("/:id/exercises/add", controller.AddExerciseToPlan)
			plans.DELETE("/:id/exercises/remove", controller.RemoveExerciseFromPlan)
			plans.PUT("/:id/exercises/reorder", controller.ReorderPlanExercises)
			plans.POST("/:id/assign-day", controller.AssignPlanToDay)
			plans.DELETE("/:id/assign-day", controller.UnassignPlanFromDay)
			plans.PUT("/:id/planned-cardio", controller.SetPlannedCardio)
			plans.PUT("/:id/planned-mobility", controller.SetPlannedMobility)
		}
		exercises := group.Group("/exercises")
		{
			exercises.GET("/all", controller.GetAllExercises)
			exercises.POST("", controller.CreateExercise)
			exercises.PUT("/:id", controller.UpdateExercise)
			exercises.PUT("/:id/cues", controller.UpdateExerciseCues)
			exercises.POST("/log", controller.LogExercise)
			exercises.POST("/add", dayOffsetMiddleware, controller.AddExerciseToWorkout)
			exercises.DELETE("/remove", dayOffsetMiddleware, controller.RemoveExerciseFromWorkout)
			exercises.DELETE("/sets/:id", controller.DeleteLoggedSet)
			exercises.GET("/progression/:id", controller.GetExerciseProgression)
		}
		logs := group.Group("/logs", dayOffsetMiddleware)
		{
			logs.GET("/today", controller.GetWorkoutToday)
			logs.GET("/month", controller.GetWorkoutMonth)
			logs.GET("/previous", controller.GetPreviousWorkout)
			logs.GET("/activity", controller.GetWorkoutActivity)
			logs.POST("/cardio", controller.UpsertCardio)
			logs.POST("/mobility/pre", controller.UpsertMobilityPre)
			logs.POST("/mobility/post", controller.UpsertMobilityPost)
			logs.PATCH("/switch-plan", controller.SwitchPlan)
		}
	}
}
