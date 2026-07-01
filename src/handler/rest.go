package handler

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	swagger "github.com/yerobalg/wealthpulse-service/docs/api"

	"github.com/yerobalg/wealthpulse-service/helper/async"
	"github.com/yerobalg/wealthpulse-service/helper/cryptolib"
	"github.com/yerobalg/wealthpulse-service/helper/logger"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/usecase"
)

type rest struct {
	http    *gin.Engine
	log     logger.Interface
	jwt     cryptolib.JWTInterface
	async   async.Interface
	usecase *usecase.Usecase
	appHost string
	appPort string
}

type InitParam struct {
	Log     logger.Interface
	JWT     cryptolib.JWTInterface
	Async   async.Interface
	Usecase *usecase.Usecase
	AppHost string
	AppPort string
}

var once = sync.Once{}

func Init(param InitParam) *rest {
	r := &rest{}

	// Initialize server with graceful shutdown
	once.Do(func() {
		gin.SetMode(gin.ReleaseMode)

		r.http = gin.New()
		r.log = param.Log
		r.jwt = param.JWT
		r.async = param.Async
		r.usecase = param.Usecase
		r.appHost = param.AppHost
		r.appPort = param.AppPort

		r.RegisterMiddlewareAndRoutes()
	})

	return r
}

func (r *rest) RegisterMiddlewareAndRoutes() {
	// Global middleware
	r.http.Use(r.CorsMiddleware())
	r.http.Use(gin.Recovery())
	r.http.Use(r.SetTimeout)
	r.http.Use(r.AddFieldsToContext)

	r.setupSwagger()

	r.http.LoadHTMLFiles("etc/template/login.html")

	// Public routes
	r.http.GET("/ping", r.Ping)
	r.http.GET("/docs/api/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.http.GET("/template/login.html", r.LoginPage)
	r.http.POST("/user/login", r.Login)

	// Protected routes
	authorized := r.http.Group("", r.Authorization())

	// User management
	authorized.GET("/user/profile", r.GetProfile)
	authorized.GET("/user", r.AuthorizePermission(entity.PermissionManageUser), r.GetListUser)
	authorized.POST("/user", r.AuthorizePermission(entity.PermissionManageUser), r.CreateUser)
	authorized.PUT("/user/:id", r.AuthorizePermission(entity.PermissionManageUser), r.UpdateUser)
}

func (r *rest) setupSwagger() {
	swagger.SwaggerInfo.Host = r.appHost + ":" + r.appPort
	swagger.SwaggerInfo.Schemes = []string{"https", "http"}
}

// @Summary Health Check
// @Description Check if the server is running
// @Tags Server
// @Produce json
// @Success 200 string example="PONG!!"
// @Router /ping [GET]
func (r *rest) Ping(c *gin.Context) {
	r.SuccessResponse(c, "PONG!!", nil, nil)
}

func (r *rest) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"loginPath": fmt.Sprintf("http://%s:%s/user/login", r.appHost, r.appPort),
	})
}

func (r *rest) Run() {
	/*
		Create context that listens for the interrupt signal from the OS.
		This will allow us to gracefully shutdown the server.
	*/
	c := context.Background()
	ctx, stop := signal.NotifyContext(c, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	port := ":8080"
	if r.appPort != "" {
		port = ":" + r.appPort
	}
	server := &http.Server{
		Addr:              port,
		Handler:           r.http,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Run the server in a goroutine so that it doesn't block the graceful shutdown handling below

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.log.Error(ctx, err.Error())
		}
	}()

	r.log.Info(context.Background(), "Server is running on port "+r.appPort)

	// Block until we receive our signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	r.log.Info(context.Background(), "Shutting down server...")

	// Create a deadline to wait for.
	quitCtx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()
	if err := server.Shutdown(quitCtx); err != nil {
		r.log.Fatal(quitCtx, fmt.Sprintf("Server Shutdown error: %s", err.Error()))
	}

	r.log.Info(context.Background(), "Server gracefully stopped")
}
