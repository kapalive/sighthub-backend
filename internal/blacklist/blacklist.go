package blacklist

import (
	"sync"
	"time"
)

// Структура для хранения данных в черном списке
type entry struct {
	expiresAt time.Time
}

var (
	mu        sync.RWMutex
	blacklist = make(map[string]entry)
)

// Добавить токен (например, его сырой вид или jti) в чёрный список
func AddToken(token string, expiresAt time.Time) {
	mu.Lock()
	defer mu.Unlock()

	if time.Now().Before(expiresAt) {
		blacklist[token] = entry{expiresAt: expiresAt}
	}
}

// Проверить, находится ли токен в чёрном списке
func IsBlacklisted(token string) bool {
	mu.RLock()
	defer mu.RUnlock()

	e, ok := blacklist[token]
	if !ok {
		return false
	}
	// если срок истёк — считаем, что его нет, но желательно почистить мапу периодически
	return time.Now().Before(e.expiresAt)
}

// Запуск горутины для чистки «протухших» токенов (если не хочешь стороний go-cache)
func StartCleaner(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			mu.Lock()
			for k, e := range blacklist {
				if time.Now().After(e.expiresAt) {
					delete(blacklist, k)
				}
			}
			mu.Unlock()
		}
	}()
}
