package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// Глобальный клиент Redis
var RDB *redis.Client

// StartRedis запускает Redis как процесс
func StartRedis() {
	cmd := exec.Command("redis-server")
	cmd.Stdout = os.Stdout // Вывод в терминал
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal("Redis startup error:", err)
	}
	fmt.Println("Redis is running in the background")
	time.Sleep(2 * time.Second) // Ждём пару секунд, чтобы Redis полностью запустился
}

// InitRedisClient инициализирует клиент Redis
func InitRedisClient() {
	RDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})
	fmt.Println("Redis client initialized")
}

// SetBlockUser блокирует пользователя на `duration`
func SetBlockUser(ctx context.Context, username string, duration time.Duration) error {
	return RDB.Set(ctx, "block:"+username, "1", duration).Err()
}

// IsUserBlocked проверяет, есть ли блокировка
func IsUserBlocked(ctx context.Context, username string) (bool, error) {
	exists, err := RDB.Exists(ctx, "block:"+username).Result()
	return exists > 0, err
}

// ClearBlockUser снимает блокировку
func ClearBlockUser(ctx context.Context, username string) error {
	return RDB.Del(ctx, "block:"+username).Err()
}

// SetJSON сохраняет произвольную структуру как JSON с TTL
func SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, key, data, ttl).Err()
}

// GetJSON извлекает JSON-значение в переданную структуру
func GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := RDB.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// DelKey удаляет ключ
func DelKey(ctx context.Context, key string) error {
	return RDB.Del(ctx, key).Err()
}

// SetString сохраняет строковое значение с TTL
func SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return RDB.Set(ctx, key, value, ttl).Err()
}

// GetString извлекает строковое значение
func GetString(ctx context.Context, key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// StartDailyCachePurge запускает очистку ключей по префиксу каждую ночь
func StartDailyCachePurge(prefix string) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next))

			ctx := context.Background()
			iter := RDB.Scan(ctx, 0, prefix+"*", 0).Iterator()
			for iter.Next(ctx) {
				_ = RDB.Del(ctx, iter.Val()).Err()
			}
		}
	}()
}

func Del(ctx context.Context, key string) error {
	return RDB.Del(ctx, key).Err()
}
