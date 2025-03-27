package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupRouter(cartHandler *CartHandler) *gin.Engine {
	router := gin.Default()

	router.Use(gin.Recovery())
	router.Use(gin.ErrorLogger())

	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	cartGroup := router.Group("carts")
	cartGroup.GET("/:owner_id", cartHandler.GetCart)
	cartGroup.POST("/:owner_id", cartHandler.AddItem)
	cartGroup.DELETE("/:owner_id/:product_id", cartHandler.DeleteItem)

	return router
}
