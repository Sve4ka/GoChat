package delivery

import (
	"backend/docs"
	"backend/internal/delivery/middleware"
	"backend/internal/delivery/routers"
	"backend/pkg/postgres"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Start(db *postgres.Pg) {
	r := gin.Default()
	r.ForwardedByClientIP = true
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	middlewareStruct := middleware.InitMiddleware()
	r.Use(middlewareStruct.CORSMiddleware())

	routers.InitRouting(r, db, middlewareStruct)

	if err := r.Run("0.0.0.0:8080"); err != nil {
		panic(fmt.Sprintf("error running client: %v", err.Error()))
	}
}
