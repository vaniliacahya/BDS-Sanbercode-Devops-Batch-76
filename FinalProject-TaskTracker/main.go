package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os" // Tambahkan package os untuk membaca env
	"strconv"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv" // Tambahkan library godotenv
	"gorm.io/gorm"
)

type Todo struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Title   string `json:"title"`
	Status  string `json:"status" gorm:"default:'todo'"`
	DueDate string `json:"due_date"`
	DueTime string `json:"due_time"`
}

var db *gorm.DB

func initDB() {
	// Membaca nama database dari .env, jika kosong gunakan default
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "kanban.db"
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database:", err)
	}
	db.AutoMigrate(&Todo{})
}

func handleTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		var todos []Todo
		db.Order("due_date asc, due_time asc").Find(&todos)
		json.NewEncoder(w).Encode(todos)

	case "POST":
		var todo Todo
		if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if todo.Status == "" {
			todo.Status = "todo"
		}
		db.Create(&todo)
		json.NewEncoder(w).Encode(todo)
	}
}

func handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var reqData struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Model(&Todo{}).Where("id = ?", reqData.ID).Update("status", reqData.Status)
	w.WriteHeader(http.StatusOK)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	db.Delete(&Todo{}, id)
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Load file .env jika ada (biasanya dipakai saat di lokal)
	// Saat di VPS, kita bisa mengabaikan file .env dan langsung mengesetnya di Systemd
	err := godotenv.Load()
	if err != nil {
		log.Println("Catatan: File .env tidak ditemukan, menggunakan environment variable dari sistem.")
	}

	initDB()

	http.HandleFunc("/api/todos", handleTodos)
	http.HandleFunc("/api/todos/update", handleUpdateStatus)
	http.HandleFunc("/api/todos/delete", handleDelete)

	// Mengambil port dari .env, default ke 8080 jika tidak ada
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Backend ProTask berjalan di port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}