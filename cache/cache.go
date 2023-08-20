package cache

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"rest-gateway/conf"
	"rest-gateway/logger"
	"sync"
)

type Message struct {
	ID           uint   `gorm:"primary key;autoIncrement" json:"id"`
	StationName  string `json:"-"`
	ConsumerName string `json:"-"`
	Username     string `json:"-"`
	Data         string `json:"data"`
}

var messageCacheLock sync.Mutex

type MessageCache interface {
	GetMessages(stationName, consumerName string, batchSize int) ([]Message, error)
	GetMessageById(stationName, consumerName string, id uint) (*Message, error)
	AddMessage(message *Message) error
	RemoveMessage(message *Message) error
}

func New(config conf.Configuration, log *logger.Logger) MessageCache {
	if config.USE_DB_CACHE {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%v TimeZone=%s",
			config.DB_HOST, config.DB_USER, config.DB_PASSWORD, config.DB_NAME, config.DB_PORT, config.DB_SSLMODE, config.DB_TIME_ZONE)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Errorf("MessageCache.New - %s", err.Error())
			log.Noticef("Defaulting to in-memory cache")
		} else {
			repo := Repository{db}
			err := repo.MigrateMessages()
			if err != nil {
				log.Errorf("MessageCache.New - %s", err.Error())
			}
			return repo
		}
	}
	return InMemoryCache{map[string]map[string][]Message{}}
}
