package core


type Store struct {
	templates map[string]Template
}

func NewStore() *Store {
	return &Store{
		templates: make(map[string]Template),
	}
}

func (s *Store) AddTemplate(template Template) error {
	// TODO: add mutex lock
	s.templates[template.Name] = template
	return nil
}

func (s *Store) TemplateExists(name string) bool {
	// TODO: add mutex lock
	_ , exists := s.templates[name]
	return exists
}