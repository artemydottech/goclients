package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os" // ← добавить
	"strconv"
	"unicode/utf8"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Не найден .env файл. Пожалуйста, создайте его используя .env.example: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"  
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"  
	}

	service, err := NewService(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer service.DB.Close()

	http.HandleFunc("POST /users", service.CreateUser)
	http.HandleFunc("GET /users", service.GetUsersList)
	http.HandleFunc("GET /users/", service.GetUserByID)

	log.Printf("Сервер на :%s", port) 
	log.Fatal(http.ListenAndServe(port, nil))
}


type Service struct {
	DB *sql.DB
}

func NewService(dbPath string) (*Service, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL COLLATE NOCASE
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	
	return &Service{DB: db}, nil
}

func (s *Service) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}
	
	var input struct{ Name string }
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	
	if input.Name == "" || !utf8.ValidString(input.Name) || len(input.Name) > 100 {
		http.Error(w, "Invalid name", 400)
		return
	}
	
	res, err := s.DB.Exec("INSERT INTO users (name) VALUES (?)", input.Name)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (s *Service) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]
	if idStr == "" {
		http.Error(w, "Missing ID", 400)
		return
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", 400)
		return
	}
	
	var name string
	err = s.DB.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err != nil {
		http.Error(w, "User not found", 404)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": name})
}

func (s *Service) GetUsersList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT id, name FROM users ORDER BY id ASC")
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}
	defer rows.Close()
	
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			http.Error(w, "Database error", 500)
			return
		}
		users = append(users, u)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]User{"users": users})
}

