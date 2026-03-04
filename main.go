package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/artemydottech/goclients/internal/handlers"
	"github.com/artemydottech/goclients/internal/repository"
	"github.com/artemydottech/goclients/internal/service"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Конфиг .env не найден")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(db)
	userServ := service.NewUserService(userRepo)
	userHdl  := handlers.NewUserHandler(userServ)

	mux := http.NewServeMux()
	
	// handlers
	mux.HandleFunc("POST /users", userHdl.CreateUser)
	mux.HandleFunc("GET /users", userHdl.GetAllUsers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на :%s", port)
	userRepo.TestRows()
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}