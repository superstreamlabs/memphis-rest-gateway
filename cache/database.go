package cache

import "gorm.io/gorm"

type Repository struct {
	DB *gorm.DB
}

func (r Repository) GetMessageById(stationName, consumerName string, id uint) (*Message, error) {
	message := Message{}
	err := r.DB.Model(&Message{}).Find(&message, "id = ? AND station_name = ? AND consumer_name = ?", id, stationName, consumerName).Error
	return &message, err
}

func (r Repository) GetMessages(stationName, consumerName string, batchSize int) ([]Message, error) {
	messages := []Message{}
	err := r.DB.Model(&Message{}).Limit(batchSize).Find(&messages, "station_name = ? AND consumer_name = ?", stationName, consumerName).Error
	return messages, err
}

func (r Repository) AddMessage(message *Message) error {
	messageCacheLock.Lock()
	defer messageCacheLock.Unlock()
	err := r.DB.Create(message).Error
	return err
}

func (r Repository) RemoveMessage(message *Message) error {
	messageCacheLock.Lock()
	defer messageCacheLock.Unlock()
	err := r.DB.Delete(message, message.ID).Error
	return err
}

func (r Repository) MigrateMessages() error {
	err := r.DB.AutoMigrate(&Message{})
	return err
}
