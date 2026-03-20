package check

import (
	"sync"

	"github.com/google/uuid"
)

type session struct {
	Results chan Result
	Total   int
}

type Store struct {
	m sync.Map
}

func NewStore() *Store {
	return &Store{}
}

// Create allocates a buffered session and returns its ID.
// total must equal the number of checkers so goroutines never block.
func (s *Store) Create(total int) (string, *session) {
	id := uuid.NewString()
	sess := &session{
		Results: make(chan Result, total),
		Total:   total,
	}
	s.m.Store(id, sess)
	return id, sess
}

func (s *Store) Get(id string) (*session, bool) {
	val, ok := s.m.Load(id)
	if !ok {
		return nil, false
	}
	return val.(*session), true
}

func (s *Store) Delete(id string) {
	s.m.Delete(id)
}
