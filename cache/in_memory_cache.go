package cache

import (
	"errors"
	"fmt"
)

type InMemoryCache struct {
	cache map[string]map[string][]Message
}

func (m InMemoryCache) GetMessages(stationName, consumerName string, batchSize int) ([]Message, error) {
	if m.cache == nil {
		return nil, errors.New("In-Memory cache wasn't initialized")
	}
	stationMessages, ok := m.cache[stationName]
	if !ok {
		return []Message{}, nil
	}
	consumerMessages, ok := stationMessages[consumerName]
	if !ok {
		return []Message{}, nil
	}
	if batchSize < len(consumerMessages) {
		return consumerMessages[:batchSize], nil
	}
	return consumerMessages, nil
}

func (m InMemoryCache) GetMessageById(stationName, consumerName string, id uint) (*Message, error) {
	if m.cache == nil {
		return nil, errors.New("In-Memory cache wasn't initialized")
	}
	stationMessages, ok := m.cache[stationName]
	if !ok {
		return nil, fmt.Errorf("message with id = %v not found", id)
	}
	consumerMessages, ok := stationMessages[consumerName]
	if !ok {
		return nil, fmt.Errorf("message with id = %v not found", id)
	}
	for _, msg := range consumerMessages {
		if msg.ID == id {
			return &msg, nil
		}
	}
	return nil, fmt.Errorf("message with id = %v not found", id)
}

func (m InMemoryCache) AddMessage(message *Message) error {
	messageCacheLock.Lock()
	defer messageCacheLock.Unlock()
	if m.cache == nil {
		m.cache = map[string]map[string][]Message{}
	}
	if m.cache[message.StationName] == nil {
		m.cache[message.StationName] = map[string][]Message{}
	}
	if m.cache[message.StationName][message.ConsumerName] == nil {
		m.cache[message.StationName][message.ConsumerName] = []Message{}
	}
	consumerMessages := m.cache[message.StationName][message.ConsumerName]
	messageId := len(consumerMessages) + 1
	message.ID = uint(messageId)
	m.cache[message.StationName][message.ConsumerName] = append(consumerMessages, *message)
	return nil
}

func (m InMemoryCache) RemoveMessage(message *Message) error {
	if m.cache == nil {
		return errors.New("in-memory cache wasn't initialized")
	}
	stationMessages, ok := m.cache[message.StationName]
	if !ok {
		return fmt.Errorf("no cache found for station - %s", message.StationName)
	}
	consumerNameMessages, ok := stationMessages[message.ConsumerName]
	if !ok {
		return fmt.Errorf("no cache found for consumer - %s", message.StationName)
	}
	index := -1
	for i, msg := range consumerNameMessages {
		if msg.ID == message.ID {
			index = i
		}
	}
	if index != -1 {
		stationMessages[message.ConsumerName] = append(consumerNameMessages[:index], consumerNameMessages[index+1:]...)
		return nil
	}
	return errors.New("message not found in cache")
}
