package finance

import (
	"be-simpletracker/internal/core/finance/controller"

	"github.com/gin-gonic/gin"
)


func  RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	group := router.Group("/finance", authMiddleware)
	{
		group.GET("/accounts", controller.GetAllAccounts)
		group.POST("/accounts", controller.CreateAccount)
		group.GET("/categories", controller.GetAllCategories)
		group.GET("/transactions", controller.GetAllTransactions)
		group.POST("/transactions", controller.CreateTransaction)
	}
}