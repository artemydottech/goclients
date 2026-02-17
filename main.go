package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	// Вот тут магия: имя_модуля/путь_к_папке
	"github.com/artemydottech/goclients/internal/handlers"
	"github.com/artemydottech/goclients/internal/repository"
	"github.com/artemydottech/goclients/internal/service"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 1. Загружаем переменные (как и было)
	if err := godotenv.Load(); err != nil {
		log.Println("Конфиг .env не найден")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}

	// 2. Инициализируем базу данных
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Создаем таблицы (можно оставить здесь для простоты, пока проект маленький)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Собираем слои (Dependency Injection)
	// ВАЖНО: Функции New... должны быть публичными (с большой буквы)
	userRepo := repository.NewUserRepository(db)
	userServ := service.NewUserService(userRepo)
	userHdl  := handlers.NewUserHandler(userServ)

	// 4. Настраиваем маршруты
	mux := http.NewServeMux()
	
	// Теперь мы вызываем метод у созданного объекта хендлера
	mux.HandleFunc("POST /users", userHdl.CreateUser)
	// Сюда же добавишь потом GET /users и т.д.

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}