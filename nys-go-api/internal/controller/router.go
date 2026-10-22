package controller

import (
	"net/http"
	"strings"
	"time"

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
	router.Use(gin.Logger(), middleware.SlowRequests(3*time.Second), middleware.CORS(cfg.Security.InnerCallHeader), middleware.Crypto(cipher), gin.Recovery(), middleware.Auth(cfg.Security, serviceLayer.JWT, serviceLayer.Sessions), middleware.RenewLogin(serviceLayer.JWT, serviceLayer.Sessions))
	_ = router.SetTrustedProxies(nil)
	router.MaxMultipartMemory = cfg.Server.MaxUploadMB << 20

	controller := &Controller{service: serviceLayer, cipher: cipher}
	router.GET("/health", func(c *gin.Context) {
		var data any
		// 数据库连接池状态：打开、使用中、空闲的连接数，以及排队等待连接的次数和总时长。
		// “使用中”长期居高不下、等待次数一直增长，说明有连接卡住，用来排查网站变慢。
		if serviceLayer.Repo != nil {
			stats := serviceLayer.Repo.DB().Stats()
			data = gin.H{"db": gin.H{
				"maxOpen":        stats.MaxOpenConnections,
				"open":           stats.OpenConnections,
				"inUse":          stats.InUse,
				"idle":           stats.Idle,
				"waitCount":      stats.WaitCount,
				"waitMs":         stats.WaitDuration.Milliseconds(),
				"closedIdleTime": stats.MaxIdleTimeClosed,
				"closedLifetime": stats.MaxLifetimeClosed,
			}}
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Go 服务运行正常", "data": data})
	})
	staticPath := strings.TrimSuffix(cfg.Upload.PublicPrefix, "/")
	// 上传文件名是随机生成的，同一地址的内容不会变，浏览器可以长期缓存；
	// gin.Dir 第二个参数为 false 时不列出目录，避免 /uploadFile/xxx/ 暴露全部已上传文件
	uploads := router.Group(staticPath, func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Next()
	})
	uploads.StaticFS("/", gin.Dir(cfg.Upload.Directory, false))

	controller.registerSEORoutes(router)
	controller.registerPublicRoutes(router)
	controller.registerBlogRoutes(router)
	controller.registerAIRoutes(router)
	return router
}
