package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/sotremont/taskflow/internal/config"
	"github.com/sotremont/taskflow/internal/transport/middleware"
)

type Server struct {
	router         *gin.Engine
	authHandler    *AuthHandler
	teamHandler    *TeamHandler
	taskHandler    *TaskHandler
	commentHandler *CommentHandler
	config         *config.Config
}

func NewServer(
	cfg *config.Config,
	authHandler *AuthHandler,
	teamHandler *TeamHandler,
	taskHandler *TaskHandler,
	commentHandler *CommentHandler,
) *Server {
	router := gin.Default()
	
	s := &Server{
		router:         router,
		authHandler:    authHandler,
		teamHandler:    teamHandler,
		taskHandler:    taskHandler,
		commentHandler: commentHandler,
		config:         cfg,
	}
	
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.router.StaticFile("/", "./web/static/index.html")
	s.router.Static("/static", "./web/static")

	v1 := s.router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", s.authHandler.Register)
			auth.POST("/login", s.authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(s.config.JWTSecret), s.authHandler.GetMe)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(s.config.JWTSecret))
		{
			teams := protected.Group("/teams")
			{
				teams.POST("", s.teamHandler.Create)
				teams.GET("", s.teamHandler.List)
				teams.GET("/all", s.teamHandler.ListAll)
				teams.GET("/:id", s.teamHandler.Get)
				teams.DELETE("/:id", s.teamHandler.Delete)
				teams.POST("/:id/members", s.teamHandler.AddMember)
				teams.POST("/:id/join", s.teamHandler.Join)
				teams.DELETE("/:id/members/:user_id", s.teamHandler.RemoveMember)
				teams.GET("/:id/members", s.teamHandler.GetMembers)
				teams.PATCH("/:id/members/:user_id/role", s.teamHandler.UpdateMemberRole)

				teams.GET("/:id/analytics", s.taskHandler.GetAnalytics)
			}
			tasks := protected.Group("/tasks")
			{
				tasks.POST("", s.taskHandler.Create)
				tasks.GET("", s.taskHandler.List)
				tasks.GET("/:id", s.taskHandler.Get)
				tasks.PATCH("/:id", s.taskHandler.Update)
				tasks.DELETE("/:id", s.taskHandler.Delete)
				tasks.POST("/:id/assign", s.taskHandler.Assign)
				tasks.GET("/:id/history", s.taskHandler.GetHistory)
				
				// Comments
				tasks.POST("/:id/comments", s.commentHandler.Create)
				tasks.GET("/:id/comments", s.commentHandler.List)
			}

			comments := protected.Group("/comments")
			{
				comments.DELETE("/:id", s.commentHandler.Delete)
			}
		}
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
