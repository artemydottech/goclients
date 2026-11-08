package app

import (
	"database/sql"
	"net/http"

	"github.com/artemydottech/goclients/internal/handlers"
	"github.com/artemydottech/goclients/internal/repository"
	"github.com/artemydottech/goclients/internal/service"
)

type Options struct {
	DebugEndpoints bool
}

func New(db *sql.DB, opts Options) http.Handler {
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
	clientSearch := handlers.NewClientSearchHandler(service.NewClientSearchService(clientsRepo))

	timeOffRepo := repository.NewTimeOffRepository(db)

	schedulesRepo := repository.NewScheduleRepository(db)
	schedulesServ := service.NewScheduleService(schedulesRepo, employeesRepo)

	appointmentsRepo := repository.NewAppointmentRepository(db)
	appointmentsServ := service.NewAppointmentService(
		appointmentsRepo, clientsRepo, employeesRepo, servicesRepo, assignmentsRepo, schedulesServ, companiesRepo, timeOffRepo,
	)
	appointmentsHandlers := handlers.NewAppointmentHandler(appointmentsServ)
	appointmentComment := handlers.NewAppointmentCommentHandler(service.NewAppointmentCommentService(appointmentsRepo))
	appointmentSearch := handlers.NewAppointmentSearchHandler(service.NewAppointmentSearchService(appointmentsRepo))
	companyReport := handlers.NewCompanyReportHandler(service.NewCompanyReportService(appointmentsRepo, companiesRepo))
	slotsServ := service.NewSlotsService(
		schedulesServ, appointmentsRepo, employeesRepo, servicesRepo, assignmentsRepo, companiesRepo, timeOffRepo,
	)
	schedulesHandlers := handlers.NewScheduleHandler(schedulesServ, slotsServ)

	timeOffServ := service.NewTimeOffService(timeOffRepo, employeesRepo, appointmentsRepo)
	timeOffHandlers := handlers.NewTimeOffHandler(timeOffServ)

	healthHandlers := handlers.NewHealthHandler(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandlers.Live)
	mux.HandleFunc("GET /readyz", healthHandlers.Ready)

	mux.HandleFunc("POST /users", userHandlers.CreateUser)
	mux.HandleFunc("GET /users", userHandlers.GetAllUsers)
	mux.HandleFunc("GET /users/{id}", userHandlers.GetUserById)
	mux.HandleFunc("DELETE /users/{id}", userHandlers.DeleteUser)

	mux.HandleFunc("POST /companies", companiesHandlers.CreateCompany)
	mux.HandleFunc("GET /companies", companiesHandlers.GetAllCompanies)
	mux.HandleFunc("GET /companies/{id}", companiesHandlers.GetCompanyById)
	mux.HandleFunc("GET /companies/{id}/report", companyReport.Get)
	mux.HandleFunc("PUT /companies/{id}", companiesHandlers.UpdateCompany)
	mux.HandleFunc("DELETE /companies/{id}", companiesHandlers.DeleteCompany)

	mux.HandleFunc("POST /employees", employeesHandlers.CreateEmployee)
	mux.HandleFunc("GET /employees", employeesHandlers.GetAllEmployees)
	mux.HandleFunc("GET /employees/{id}", employeesHandlers.GetEmployeeById)
	mux.HandleFunc("PUT /employees/{id}", employeesHandlers.UpdateEmployee)
	mux.HandleFunc("DELETE /employees/{id}", employeesHandlers.DeleteEmployee)

	mux.HandleFunc("POST /services", servicesHandlers.CreateService)
	mux.HandleFunc("GET /services", servicesHandlers.GetAllServices)
	mux.HandleFunc("GET /services/{id}", servicesHandlers.GetServiceById)
	mux.HandleFunc("PUT /services/{id}", servicesHandlers.UpdateService)
	mux.HandleFunc("DELETE /services/{id}", servicesHandlers.DeleteService)

	mux.HandleFunc("PUT /employees/{id}/services", assignmentsHandlers.SetEmployeeServices)
	mux.HandleFunc("GET /employees/{id}/services", assignmentsHandlers.GetEmployeeServices)
	mux.HandleFunc("GET /services/{id}/employees", assignmentsHandlers.GetServiceEmployees)

	mux.HandleFunc("POST /clients", clientsHandlers.CreateClient)
	mux.HandleFunc("GET /clients", clientsHandlers.GetAllClients)
	mux.HandleFunc("GET /clients/search", clientSearch.Search)
	mux.HandleFunc("GET /clients/export", clientSearch.Export)
	mux.HandleFunc("GET /clients/{id}", clientsHandlers.GetClientById)
	mux.HandleFunc("PUT /clients/{id}", clientsHandlers.UpdateClient)
	mux.HandleFunc("DELETE /clients/{id}", clientsHandlers.DeleteClient)
	mux.HandleFunc("GET /clients/{id}/stats", appointmentsHandlers.GetClientStats)

	mux.HandleFunc("POST /appointments", appointmentsHandlers.CreateAppointment)
	mux.HandleFunc("GET /appointments", appointmentsHandlers.GetAllAppointments)
	mux.HandleFunc("GET /appointments/search", appointmentSearch.Search)
	mux.HandleFunc("GET /appointments/{id}", appointmentsHandlers.GetAppointmentById)
	mux.HandleFunc("PUT /appointments/{id}/status", appointmentsHandlers.SetStatus)
	mux.HandleFunc("PUT /appointments/{id}/comment", appointmentComment.Update)
	mux.HandleFunc("PUT /appointments/{id}/time", appointmentsHandlers.Reschedule)
	mux.HandleFunc("DELETE /appointments/{id}", appointmentsHandlers.DeleteAppointment)

	mux.HandleFunc("PUT /employees/{id}/schedule", schedulesHandlers.SetEmployeeSchedule)
	mux.HandleFunc("GET /employees/{id}/schedule", schedulesHandlers.GetEmployeeSchedule)
	mux.HandleFunc("GET /slots", schedulesHandlers.GetFreeSlots)

	mux.HandleFunc("POST /employees/{id}/time-off", timeOffHandlers.CreateTimeOff)
	mux.HandleFunc("GET /employees/{id}/time-off", timeOffHandlers.GetTimeOff)
	mux.HandleFunc("DELETE /time-off/{id}", timeOffHandlers.DeleteTimeOff)

	if opts.DebugEndpoints {
		mountDebug(mux)
	}

	return mux
}
