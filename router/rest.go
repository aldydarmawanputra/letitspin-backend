package router

import (
	"database/sql"
	"net/http"
	"strings"

	"let-it-spin/internal/auth/jwt"
	"let-it-spin/internal/auth/middleware"
	authRepository "let-it-spin/internal/auth/repository"
	"let-it-spin/internal/config"
	"let-it-spin/internal/game/blackjack"
	"let-it-spin/internal/game/dice"
	gameHandler "let-it-spin/internal/game/handler"
	gameRepository "let-it-spin/internal/game/repository"
	"let-it-spin/internal/game/roulette"
	gameService "let-it-spin/internal/game/service"
	"let-it-spin/internal/game/slot"
	"let-it-spin/internal/handler"
	"let-it-spin/internal/repository"
	walletHandler "let-it-spin/internal/wallet/handler"
	walletRepository "let-it-spin/internal/wallet/repository"
	walletService "let-it-spin/internal/wallet/service"
	"let-it-spin/internal/websocket"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	cfg := config.LoadConfig()

	jwtService := jwt.NewJWTService(cfg)

	userRepo := repository.NewUserRepository(db)
	authRepo := authRepository.NewAuthRepository(db)
	walletRepo := walletRepository.NewWalletRepository(db)
	gameRepo := gameRepository.NewGameRepository(db)
	configRepo := gameRepository.NewConfigRepository(db)
	bjRepo := blackjack.NewBlackjackRepository(db)

	wsHub := websocket.NewHub()
	go wsHub.Run()

	walletSvc := walletService.NewWalletService(walletRepo, wsHub)
	gameSvc := gameService.NewGameService(gameRepo, walletRepo, walletSvc, wsHub)

	slotEngine := slot.NewSlotEngine(configRepo)
	gameSvc.RegisterEngine(slotEngine)

	diceEngine := dice.NewDiceEngine(configRepo)
	gameSvc.RegisterEngine(diceEngine)

	rouletteEngine := roulette.NewRouletteEngine(configRepo)
	gameSvc.RegisterEngine(rouletteEngine)

	blackjackEngine := blackjack.NewBlackjackEngine(configRepo, bjRepo)
	gameSvc.RegisterEngine(blackjackEngine)

	userHandler := handler.NewUserHandler(userRepo)
	authHandler := handler.NewAuthHandler(authRepo, jwtService)
	walletHandler := walletHandler.NewWalletHandler(walletSvc)
	gameHandler := gameHandler.NewGameHandler(gameSvc)

	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/ws", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		wsHub.HandleWebSocket(c)
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("/by-email", userHandler.GetUserByEmail)
		}

		protected := v1.Group("/")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/users/:id", userHandler.GetUserByID)
			protected.PATCH("/users/:id", userHandler.PatchUser)
			protected.DELETE("/users/:id", userHandler.DeleteUser)

			protected.GET("/me", authHandler.GetMe)

			protected.GET("/wallet/balance", walletHandler.GetBalance)
			protected.POST("/wallet/deposit", walletHandler.Deposit)
			protected.POST("/wallet/withdraw", walletHandler.Withdraw)
			protected.GET("/wallet/transactions", walletHandler.GetTransactions)

			protected.GET("/games/:code/config", gameHandler.GetGameConfig)
			protected.POST("/games/:code/play", gameHandler.Play)
			protected.GET("/games/history", gameHandler.GetHistory)
		}
	}

	return r
}
