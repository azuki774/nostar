package db

import (
	"context"
	"errors"
	"nostar/internal/relay/domain"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DSN string
}

func NewGormDB(ctx context.Context, cfg Config) (*gorm.DB, error) {
	if cfg.DSN == "" {
		return nil, errors.New("empty DSN")
	}

	gdb, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		zap.S().Errorw("failed to open database", "error", err)
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		zap.S().Errorw("failed to get generic DB", "error", err)
		return nil, err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		zap.S().Errorw("database ping failed", "error", err)
		return nil, err
	}

	zap.S().Infow("database connection established")
	return gdb, nil
}

type EventStore struct {
	db *gorm.DB
}

// TODO: 具体的に実装する
func NewEventStore(db *gorm.DB) *EventStore {
	return &EventStore{
		db: db,
	}
}

func (e *EventStore) Save(ctx context.Context, evt domain.Event) error {
	zap.S().Infow("not yet implemeneted", "type", "Save")
	return nil
}

func (e *EventStore) Query(ctx context.Context, sub domain.Subscription) ([]domain.Event, error) {
	zap.S().Infow("not yet implemeneted", "type", "Query")
	return []domain.Event{}, nil
}
