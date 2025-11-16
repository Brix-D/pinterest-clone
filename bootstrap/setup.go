package bootstrap

import (
	"context"
	"encoding/json"
	"log"
	"main/utils"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func SetupApp(root string) *fiber.App {

	// .env
	envPath := path.Join(root, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Подключения
	utils.ConnectDatabase()

	app := fiber.New(fiber.Config{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ServerHeader: "Fiber",
		AppName:      os.Getenv("APP_NAME"),
	})

	// middlewares
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format:     "${locals:requestid}: ${time} ${method} ${path} - ${status} - ${latency}\n",
		TimeFormat: time.DateTime,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:          200,
		Expiration:   60 * time.Second,
		KeyGenerator: func(context *fiber.Ctx) string { return context.IP() },
	}))
	app.Use(recover.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Content-Type, Authorization, X-Requested-With",
	}))

	// routes

	// Создаем экземпляр cron
	cronInstance := cron.New()
	// Запускаем cron
	cronInstance.Start()

	// run app
	go func() {
		if err := app.Listen(":" + os.Getenv("APP_PORT")); err != nil {
			logrus.Errorf("fiber listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cronContext := cronInstance.Stop()

	select {
	case <-cronContext.Done():
	case <-ctx.Done():
	}

	shitdownError := app.Shutdown()

	if shitdownError != nil {
		logrus.Errorf("fiber shutdown error: %v", shitdownError)
	}

	return app
}
