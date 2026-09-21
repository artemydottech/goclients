package main

import (
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/artemydottech/goclients/internal/handlers"
	"github.com/artemydottech/goclients/internal/repository"
	"github.com/artemydottech/goclients/internal/service"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env config not found")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}

	db, err := repository.Open(dbPath)
	if err != nil {
		log.Fatal("db connection failed: ", err)
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

	assignmentsRepo := repository.NewAssignmentRepository(db)
	assignmentsServ := service.NewAssignmentService(assignmentsRepo, employeesRepo, servicesRepo)
	assignmentsHandlers := handlers.NewAssignmentHandler(assignmentsServ)

	clientsRepo := repository.NewClientRepository(db)
	clientsServ := service.NewClientService(clientsRepo)
	clientsHandlers := handlers.NewClientHandler(clientsServ)

	timeOffRepo := repository.NewTimeOffRepository(db)

	schedulesRepo := repository.NewScheduleRepository(db)
	schedulesServ := service.NewScheduleService(schedulesRepo, employeesRepo)

	appointmentsRepo := repository.NewAppointmentRepository(db)
	appointmentsServ := service.NewAppointmentService(
		appointmentsRepo, clientsRepo, employeesRepo, servicesRepo, assignmentsRepo, schedulesServ, companiesRepo, timeOffRepo,
	)
	appointmentsHandlers := handlers.NewAppointmentHandler(appointmentsServ)
	slotsServ := service.NewSlotsService(
		schedulesServ, appointmentsRepo, employeesRepo, servicesRepo, assignmentsRepo, companiesRepo, timeOffRepo,
	)
	schedulesHandlers := handlers.NewScheduleHandler(schedulesServ, slotsServ)

	timeOffServ := service.NewTimeOffService(timeOffRepo, employeesRepo, appointmentsRepo)
	timeOffHandlers := handlers.NewTimeOffHandler(timeOffServ)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", userHandlers.CreateUser)
	mux.HandleFunc("GET /users", userHandlers.GetAllUsers)
	mux.HandleFunc("GET /users/{id}", userHandlers.GetUserById)
	mux.HandleFunc("DELETE /users/{id}", userHandlers.DeleteUser)

	mux.HandleFunc("POST /companies", companiesHandlers.CreateCompany)
	mux.HandleFunc("GET /companies", companiesHandlers.GetAllCompanies)
	mux.HandleFunc("GET /companies/{id}", companiesHandlers.GetCompanyById)
	mux.HandleFunc("DELETE /companies/{id}", companiesHandlers.DeleteCompany)

	mux.HandleFunc("POST /employees", employeesHandlers.CreateEmployee)
	mux.HandleFunc("GET /employees", employeesHandlers.GetAllEmployees)
	mux.HandleFunc("GET /employees/{id}", employeesHandlers.GetEmployeeById)
	mux.HandleFunc("DELETE /employees/{id}", employeesHandlers.DeleteEmployee)

	mux.HandleFunc("POST /services", servicesHandlers.CreateService)
	mux.HandleFunc("GET /services", servicesHandlers.GetAllServices)
	mux.HandleFunc("GET /services/{id}", servicesHandlers.GetServiceById)
	mux.HandleFunc("DELETE /services/{id}", servicesHandlers.DeleteService)

	mux.HandleFunc("PUT /employees/{id}/services", assignmentsHandlers.SetEmployeeServices)
	mux.HandleFunc("GET /employees/{id}/services", assignmentsHandlers.GetEmployeeServices)
	mux.HandleFunc("GET /services/{id}/employees", assignmentsHandlers.GetServiceEmployees)

	mux.HandleFunc("POST /clients", clientsHandlers.CreateClient)
	mux.HandleFunc("GET /clients", clientsHandlers.GetAllClients)
	mux.HandleFunc("GET /clients/{id}", clientsHandlers.GetClientById)
	mux.HandleFunc("DELETE /clients/{id}", clientsHandlers.DeleteClient)
	mux.HandleFunc("GET /clients/{id}/stats", appointmentsHandlers.GetClientStats)

	mux.HandleFunc("POST /appointments", appointmentsHandlers.CreateAppointment)
	mux.HandleFunc("GET /appointments", appointmentsHandlers.GetAllAppointments)
	mux.HandleFunc("GET /appointments/{id}", appointmentsHandlers.GetAppointmentById)
	mux.HandleFunc("PUT /appointments/{id}/status", appointmentsHandlers.SetStatus)
	mux.HandleFunc("PUT /appointments/{id}/time", appointmentsHandlers.Reschedule)
	mux.HandleFunc("DELETE /appointments/{id}", appointmentsHandlers.DeleteAppointment)

	mux.HandleFunc("PUT /employees/{id}/schedule", schedulesHandlers.SetEmployeeSchedule)
	mux.HandleFunc("GET /employees/{id}/schedule", schedulesHandlers.GetEmployeeSchedule)
	mux.HandleFunc("GET /slots", schedulesHandlers.GetFreeSlots)

	mux.HandleFunc("POST /employees/{id}/time-off", timeOffHandlers.CreateTimeOff)
	mux.HandleFunc("GET /employees/{id}/time-off", timeOffHandlers.GetTimeOff)
	mux.HandleFunc("DELETE /time-off/{id}", timeOffHandlers.DeleteTimeOff)

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

	log.Printf("server listening on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
