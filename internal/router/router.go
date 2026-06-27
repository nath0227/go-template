package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httphandler "github.com/your-org/service-name/internal/handler/http"
	"github.com/your-org/service-name/pkg/httpserver/middleware"
)

func New(product *httphandler.ProductHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery(), middleware.TID(), middleware.Logger(), middleware.CORS())

	r.GET("/system/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	v1 := r.Group("/api/v1")
	{
		products := v1.Group("/products")
		products.POST("", product.Create)
		products.GET("", product.List)
		products.GET("/:id", product.GetByID)
		products.PUT("/:id", product.Update)
		products.DELETE("/:id", product.Delete)
	}

	return r
}
