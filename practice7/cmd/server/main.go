package main

import (
	"log"
	"os"

	v1 "practice7/internal/controller/http/v1"
	"practice7/internal/entity"
	"practice7/internal/usecase"
	"practice7/internal/usecase/repo"
	"practice7/pkg/logger"
	"practice7/pkg/mail"
	"practice7/pkg/postgres"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=practice7 port=5432 sslmode=disable"
	}
	pg, err := postgres.New(dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := pg.Conn.AutoMigrate(&entity.User{}); err != nil {
		log.Fatal(err)
	}

	l := logger.New()
	m := mail.NewFromEnv()
	ur := repo.NewUserRepo(pg)
	uc := usecase.NewUserUseCase(ur, m)

	r := gin.Default()
	v1.RegisterRoutes(r, uc, l)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("listening %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
