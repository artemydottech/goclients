package main

import (
	"log"
	"net/http"
	"os"
	"time"

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

	servicesRepo := repository.NewServiceRepository(db)
	catalogServ := service.NewCatalogService(servicesRepo)
	servicesHandlers := handlers.NewServiceHandler(catalogServ)

	mux := http.NewServeMux()

	// handlers

	//users
	mux.HandleFunc("POST /users", userHandlers.CreateUser)
	mux.HandleFunc("GET /users", userHandlers.GetAllUsers)
	mux.HandleFunc("GET /users/{id}", userHandlers.GetUserById)
	mux.HandleFunc("DELETE /users/{id}", userHandlers.DeleteUser)

	//companies
	mux.HandleFunc("POST /companies", companiesHandlers.CreateCompany)
	mux.HandleFunc("GET /companies", companiesHandlers.GetAllCompanies)
	mux.HandleFunc("GET /companies/{id}", companiesHandlers.GetCompanyById)
	mux.HandleFunc("DELETE /companies/{id}", companiesHandlers.DeleteCompany)

	//employees
	mux.HandleFunc("POST /employees", employeesHandlers.CreateEmployee)
	mux.HandleFunc("GET /employees", employeesHandlers.GetAllEmployees)
	mux.HandleFunc("GET /employees/{id}", employeesHandlers.GetEmployeeById)
	mux.HandleFunc("DELETE /employees/{id}", employeesHandlers.DeleteEmployee)

	//services
	mux.HandleFunc("POST /services", servicesHandlers.CreateService)
	mux.HandleFunc("GET /services", servicesHandlers.GetAllServices)
	mux.HandleFunc("GET /services/{id}", servicesHandlers.GetServiceById)
	mux.HandleFunc("DELETE /services/{id}", servicesHandlers.DeleteService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Сервер запущен на :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
