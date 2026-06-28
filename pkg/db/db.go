package db

// Store is the unified database handle.
// GORM is non-nil when Backend is "gorm" or "both".
// PGX is non-nil when Backend is "pgx" or "both".
type Store struct {
	GORM *DB
	PGX  *PGXPool
}

// New initializes DB connections based on cfg.Backend.
func New(cfg DBConfig) (*Store, error) {
	var (
		s   Store
		err error
	)
	switch cfg.Backend {
	case BackendPGX:
		s.PGX, err = NewPGX(cfg)
	case BackendBoth:
		if s.GORM, err = NewGORM(cfg); err != nil {
			return nil, err
		}
		if s.PGX, err = NewPGX(cfg); err != nil {
			(&Store{GORM: s.GORM}).Close()
			return nil, err
		}
	default: // BackendGORM
		s.GORM, err = NewGORM(cfg)
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Close releases all open connections.
func (s *Store) Close() {
	if s.GORM != nil {
		if sqlDB, err := s.GORM.Writer.DB(); err == nil {
			_ = sqlDB.Close()
		}
		if s.GORM.Reader != s.GORM.Writer {
			if sqlDB, err := s.GORM.Reader.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	}
	if s.PGX != nil {
		s.PGX.Writer.Close()
		if s.PGX.Reader != s.PGX.Writer {
			s.PGX.Reader.Close()
		}
	}
}
