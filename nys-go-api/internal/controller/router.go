package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/config"
	"nys-go-api/internal/middleware"
	"nys-go-api/internal/security"
	"nys-go-api/internal/service"
)

type Controller struct {
	service *service.Service
	cipher  *security.AESCipher
}

func NewRouter(cfg *config.Config, serviceLayer *service.Service, cipher *security.AESCipher) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(gin.Logger(), middleware.CORS(cfg.Security.InnerCallHeader), middleware.Crypto(cipher), gin.Recovery(), middleware.Auth(cfg.Security, serviceLayer.JWT))
	_ = router.SetTrustedProxies(nil)
	router.MaxMultipartMemory = cfg.Server.MaxUploadMB << 20

	controller := &Controller{service: serviceLayer, cipher: cipher}
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Go 服务运行正常", "data": nil})
	})
	staticPath := strings.TrimSuffix(cfg.Upload.PublicPrefix, "/")
	router.StaticFS(staticPath, http.Dir(cfg.Upload.Directory))

	controller.registerSEORoutes(router)
	controller.registerPublicRoutes(router)
	controller.registerBlogRoutes(router)
	controller.registerAIRoutes(router)
	return router
}
