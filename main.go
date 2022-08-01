package main

import (
	"database/sql"
	"fmt"
	"os"
	"target/onboarding-assignment/http/handlers"
	"target/onboarding-assignment/repository"
	"target/onboarding-assignment/server"
	"target/onboarding-assignment/services"
	"time"

	migrate "github.com/rubenv/sql-migrate"

	"github.com/gin-gonic/gin"
)

func main() {
	var conn *sql.DB
	var err error

	//Initializing database connection, attempting five times and exiting if not successfull, with 5 seconds waiting between attempts
	fmt.Println("Connecting to database")
	for i := 0; i < 5; i++ {
		conn, _ = repository.ConnectToDatabase()
		err = conn.Ping()
		if err != nil {
			fmt.Println(err)
			if i == 4 {
				os.Exit(7)
			}
			time.Sleep(5 * time.Second)
		} else {
			fmt.Println("Connection to db success")

			file, err := os.Create("/var/ready")
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("File created successfully")
			file.Close()

			break
		}
	}
	defer conn.Close()

	fmt.Println("Defining migrations path")
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations/postgres",
	}

	fmt.Println("Starting migrations")
	n, err := migrate.Exec(conn, "postgres", migrations, migrate.Up)
	if err != nil {
		fmt.Println("ERROR WHEN MIGRATING UP", err)
	}
	fmt.Printf("Applied %d migrations!\n", n)

	//Initializing dependency chain
	fmt.Println("Initializing dependency chain")
	userRepo := repository.UserRepository{Db: conn}
	userService := services.UserService{Repo: &userRepo}
	userHandler := handlers.UserHandler{UserSvc: &userService}
	server := server.Server{HTTPHandler: &userHandler, Router: gin.Default(), Port: ":8080"}
	server.InitRoutes()
	server.Run()

}
