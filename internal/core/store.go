package core

import "sync"

type Store struct {
	templates map[string]Template
	pools     map[string]*Pool
	mutex     sync.RWMutex
	Metrics   *Metrics
}

type Metrics struct {
	allocations int64
	releases    int64
	timeouts    int64
	mutex       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		templates: make(map[string]Template),
		pools:     make(map[string]*Pool),
		Metrics:   &Metrics{},
	}
}

func (s *Store) CreateTemplate(template Template) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.templates[template.Name] = template
	return nil
}

func (s *Store) TemplateExists(name string) (Template, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	template, exists := s.templates[name]
	return template, exists
}

func (s *Store) CreatePool(pool *Pool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.pools[pool.Name] = pool
	return nil
}

func (s *Store) PoolExists(name string) (*Pool, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	pool, exists := s.pools[name]
	return pool, exists
}

func (m *Metrics) IncrementAllocations() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.allocations++
}

func (m *Metrics) IncrementReleases() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.releases++
}

func (m *Metrics) IncrementTimeouts() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.timeouts++
}

func (m *Metrics) GetMetrics() (int64, int64, int64) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.allocations, m.releases, m.timeouts
}