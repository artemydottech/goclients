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

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS companies (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        address TEXT,
        geolocation TEXT,
        schedule TEXT,
        site TEXT,
        socials TEXT,  
        logo TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(db)
	userServ := service.NewUserService(userRepo)
	userHandlers := handlers.NewUserHandler(userServ)

	companiesRepo := repository.NewCompanyRepository(db)
	companiesServ := service.NewCompanyService(companiesRepo)
	companiesHandlers := handlers.NewCompanyHandler(companiesServ)

	mux := http.NewServeMux()

	// handlers

	//users
	mux.HandleFunc("POST /users", userHandlers.CreateUser)
	mux.HandleFunc("GET /users", userHandlers.GetAllUsers)
	mux.HandleFunc("GET /users/", userHandlers.GetUserById)
	mux.HandleFunc("DELETE /users/", userHandlers.DeleteUser)

	//companies
	mux.HandleFunc("POST /companies", companiesHandlers.CreateCompany)
	mux.HandleFunc("GET /companies", companiesHandlers.GetAllCompanies)
	mux.HandleFunc("GEt /companies/", companiesHandlers.GetCompanyById)
	mux.HandleFunc("DELETE /companies/", companiesHandlers.DeleteCompany)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на :%s", port)
	userRepo.TestRows()
	companiesRepo.TestRows()
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
