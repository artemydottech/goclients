package main

import (
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

	db, err := repository.Open(dbPath)
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	defer db.Close()

	if err := repository.Migrate(db); err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(db)
	userServ := service.NewUserService(userRepo)
	userHandlers := handlers.NewUserHandler(userServ)

	companiesRepo := repository.NewCompanyRepository(db)
	companiesServ := service.NewCompanyService(companiesRepo)
	companiesHandlers := handlers.NewCompanyHandler(companiesServ)

	employeesRepo := repository.NewEmployeeRepository(db)
	employeesServ := service.NewEmployeeService(employeesRepo)
	employeesHandlers := handlers.NewEmployeeHandler(employeesServ)

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
	mux.HandleFunc("GET /companies/", companiesHandlers.GetCompanyById)
	mux.HandleFunc("DELETE /companies/", companiesHandlers.DeleteCompany)

	//employees
	mux.HandleFunc("POST /employees", employeesHandlers.CreateEmployee)
	mux.HandleFunc("GET /employees", employeesHandlers.GetAllEmployees)
	mux.HandleFunc("GET /employees/", employeesHandlers.GetEmployeeById)
	mux.HandleFunc("DELETE /employees/", employeesHandlers.DeleteEmployee)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
