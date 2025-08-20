package core

import "sync"

type Store struct {
	templates map[string]Template
	mutex     sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		templates: make(map[string]Template),
	}
}

func (s *Store) AddTemplate(template Template) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.templates[template.Name] = template
	return nil
}

func (s *Store) TemplateExists(name string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.templates[name]
	return exists
}