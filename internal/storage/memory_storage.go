package storage

import (
	"sync"
	"time"
)

type MemoryStorage struct {
	messages  map[uint64]*Message
	mu        sync.Mutex
	currentID uint64
}

func New() *MemoryStorage {
	return &MemoryStorage{
		messages:  make(map[uint64]*Message),
		currentID: 1,
	}
}

func (s *MemoryStorage) Save(message *Message) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	message.ID = s.currentID
	message.CreatedAt = time.Now()
	s.currentID++

	s.messages[message.ID] = message

	return message, nil
}

func (s *MemoryStorage) GetAll() ([]*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages := make([]*Message, 0, len(s.messages))
	for _, message := range s.messages {
		messages = append(messages, message)
	}

	return messages, nil
}

func (s *MemoryStorage) Delete(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.messages, id)

	return nil
}
