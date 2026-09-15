package storage

import "time"

type Storage interface {
	Save(message *Message) (*Message, error)
	GetAll() ([]*Message, error)
	Delete(id uint64) error
}

type Message struct {
	CreatedAt time.Time `json:"created_at,omitzero"`
	Message   string    `json:"message"`
	ID        uint64    `json:"id,omitempty"`
}
