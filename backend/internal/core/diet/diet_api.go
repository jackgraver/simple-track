package diet

import (
	"be-simpletracker/internal/core/diet/controller"
	"be-simpletracker/internal/utils"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	dayOffsetMiddleware := utils.DayOffsetMiddleware()

	group := router.Group("/diet", authMiddleware)
	{
		plans := group.Group("/plans")
		{
			plans.GET("/plan/all", controller.GetAllPlans)
			plans.PUT("/plan/:id", controller.PutPlanMacros)
		}
		logs := group.Group("/logs")
		{
			logs.GET("/today", dayOffsetMiddleware, controller.GetMealPlanToday)
			logs.GET("/week", controller.GetMealPlanWeek)
			logs.GET("/month-planned-summary", controller.GetMonthPlannedSummary)
			logs.GET("/month", controller.GetMealPlanMonth)
			logs.GET("/day/:id", controller.GetMealPlanDay)
			logs.GET("/goals/today", controller.GetGoalsToday)
		}
		foods := group.Group("/foods")
		{
			foods.POST("", controller.PostFood)
		}
		meals := group.Group("/meals")
		{
			meals.GET("/food/all", controller.GetAllFoods)
			meals.POST("/composite-food/new", controller.PostNewCompositeFood)
			meals.GET("/meal/all", controller.GetAllMeals)
			meals.GET("/saved-meal/all", controller.GetAllSavedMeals)
			meals.GET("/saved-meal/:id", controller.GetSavedMeal)
			meals.POST("/saved-meal/new", controller.PostNewSavedMeal)
			meals.PUT("/saved-meal/:id", controller.PutSavedMeal)
			meals.DELETE("/saved-meal/:id", controller.DeleteSavedMeal)
			meals.GET("/meal/:id", controller.GetMeal)
			meals.POST("/quick-log", controller.PostQuickLog)
			meals.POST("/meal/new", controller.PostNewMeal)
			meals.POST("/meal/log-planned", controller.PostLogPlanned)
			meals.POST("/meal/logedited", controller.PostLogEdited)
			meals.POST("/meal/editlogged", controller.PostEditLogged)
			meals.DELETE("/meal/logged", controller.DeleteLoggedMeal)
			meals.POST("/planned/from-saved", controller.PostPlannedFromSaved)
			meals.POST("/planned/reorder", controller.PostPlannedReorder)
			meals.DELETE("/planned", controller.DeletePlannedMeal)
		}
	}
}
