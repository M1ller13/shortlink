package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	db  *sql.DB
	rdb *redis.Client
	ctx = context.Background()
)

type Link struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
}

func main() {
	// Загрузка .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Подключение к PostgreSQL
	db, err = sql.Open("postgres", os.Getenv("POSTGRES_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Инициализация Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	// Создание таблицы
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS links (
		id SERIAL PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_code TEXT UNIQUE NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Настройка роутера Gin
	r := gin.Default()

	// Эндпоинты
	r.POST("/shorten", shortenURL)
	r.GET("/:code", redirectURL)

	// Запуск сервера
	r.Run(":8080")
}

// Генерация короткого кода
func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Обработчик создания короткой ссылки
func shortenURL(c *gin.Context) {
	var input struct {
		URL string `json:"url"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Проверка кэша Redis
	cachedURL, err := rdb.Get(ctx, input.URL).Result()
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"short_url": cachedURL})
		return
	}

	// Генерация кода
	shortCode := generateShortCode()

	// Сохранение в PostgreSQL
	_, err = db.Exec("INSERT INTO links (original_url, short_code) VALUES ($1, $2)", input.URL, shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	// Сохранение в Redis
	shortURL := fmt.Sprintf("http://localhost:8080/%s", shortCode)
	err = rdb.Set(ctx, input.URL, shortURL, 24*time.Hour).Err()
	if err != nil {
		log.Println("Redis cache error:", err)
	}

	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
}

// Редирект по короткой ссылке
func redirectURL(c *gin.Context) {
	code := c.Param("code")

	// Проверка кэша Redis
	cachedURL, err := rdb.Get(ctx, code).Result()
	if err == nil {
		c.Redirect(http.StatusMovedPermanently, cachedURL)
		return
	}

	// Поиск в PostgreSQL
	var originalURL string
	err = db.QueryRow("SELECT original_url FROM links WHERE short_code = $1", code).Scan(&originalURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Кэширование
	err = rdb.Set(ctx, code, originalURL, 24*time.Hour).Err()
	if err != nil {
		log.Println("Redis cache error:", err)
	}

	c.Redirect(http.StatusMovedPermanently, originalURL)
}
