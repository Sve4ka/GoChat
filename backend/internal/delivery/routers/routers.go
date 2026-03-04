package routers

import (
	"backend/internal/delivery/middleware"
	"backend/pkg/postgres"
	"github.com/gin-gonic/gin"
)

func InitRouting(r *gin.Engine, db *postgres.Pg, middlewareStruct middleware.Middleware) {
	_ = RegisterUserRouter(r, db)
	_ = RegisterChatRouter(r, db)
}
