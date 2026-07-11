package check

import (
	"sync"

	"github.com/google/uuid"
)

type Session struct {
	Results chan Result
	Total   int
}

type Store struct {
	m sync.Map
}

func NewStore() *Store {
	return &Store{}
}

// Create allocates a buffered Session and returns its ID.
// total must equal the number of checkers so goroutines never block.
func (s *Store) Create(total int) (string, *Session) {
	return s.CreateWithID(uuid.NewString(), total)
}

// CreateWithID allocates a buffered Session under a caller-provided ID (e.g. a
// brand-kit ID) so the same identifier can be used for streaming and later
// retrieval. total must equal the number of results so senders never block.
func (s *Store) CreateWithID(id string, total int) (string, *Session) {
	sess := &Session{
		Results: make(chan Result, total),
		Total:   total,
	}
	s.m.Store(id, sess)
	return id, sess
}

func (s *Store) Get(id string) (*Session, bool) {
	val, ok := s.m.Load(id)
	if !ok {
		return nil, false
	}
	return val.(*Session), true
}

func (s *Store) Delete(id string) {
	s.m.Delete(id)
}
