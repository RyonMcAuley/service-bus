package store

func (s *SqliteStore) migrate() error {
	_, err := s.db.Exec(schemaMigration)
	return err
}
