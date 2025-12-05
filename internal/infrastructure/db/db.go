package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// EventModel is the GORM model for storing Nostr events
type EventModel struct {
	ID        string `gorm:"primaryKey;size:64"`
	Pubkey    string `gorm:"index;size:64;not null"`
	Sig       string `gorm:"size:128;not null"`
	CreatedAt int64  `gorm:"index;not null"`
	Kind      int    `gorm:"index;not null"`
	Tags      string `gorm:"type:jsonb"`
	Content   string `gorm:"type:text"`
}

func (EventModel) TableName() string {
	return "events"
}

// toModel converts domain.Event to EventModel for database storage
func toModel(evt domain.Event) (EventModel, error) {
	tagsJSON, err := json.Marshal(evt.Tags)
	if err != nil {
		return EventModel{}, fmt.Errorf("failed to marshal tags: %w", err)
	}

	return EventModel{
		ID:        evt.ID,
		Pubkey:    evt.PubKey,
		Sig:       evt.Signature,
		CreatedAt: evt.CreatedAt,
		Kind:      evt.Kind,
		Tags:      string(tagsJSON),
		Content:   evt.Content,
	}, nil
}

// toDomain converts EventModel to domain.Event
// こちら向きを使うようになったらコメントアウトを外す
func toDomain(model EventModel) (domain.Event, error) {
	var tags [][]string
	if err := json.Unmarshal([]byte(model.Tags), &tags); err != nil {
		return domain.Event{}, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return domain.Event{
		ID:        model.ID,
		PubKey:    model.Pubkey,
		Signature: model.Sig,
		CreatedAt: model.CreatedAt,
		Kind:      model.Kind,
		Tags:      tags,
		Content:   model.Content,
	}, nil
}

type EventStore struct {
	db *gorm.DB
}

func NewEventStore(db *gorm.DB) *EventStore {
	return &EventStore{
		db: db,
	}
}

func (e *EventStore) Save(ctx context.Context, evt domain.Event) error {
	model, err := toModel(evt) // domain -> DBモデルに変換
	if err != nil {
		return fmt.Errorf("failed to convert to model: %w", err)
	}

	if err := e.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

func (e *EventStore) Query(ctx context.Context, sub domain.Subscription) ([]domain.Event, error) {
	// TODO: タグ検索を実装する
	var results []domain.Event

	for _, filter := range sub.Filters {
		var models []EventModel
		query := e.db.WithContext(ctx).Model(&EventModel{})

		// IDs filter
		if len(filter.IDs) > 0 {
			query = query.Where("id IN ?", filter.IDs)
		}

		// Authors filter
		if len(filter.Authors) > 0 {
			query = query.Where("pubkey IN ?", filter.Authors)
		}

		// Kinds filter
		if len(filter.Kinds) > 0 {
			query = query.Where("kind IN ?", filter.Kinds)
		}

		// Time range filters
		if filter.Since != nil {
			query = query.Where("created_at >= ?", *filter.Since)
		}
		if filter.Until != nil {
			query = query.Where("created_at <= ?", *filter.Until)
		}

		// Limit (各フィルタに適用)
		if filter.Limit != nil {
			query = query.Limit(*filter.Limit)
		}

		if err := query.Find(&models).Error; err != nil {
			return nil, err
		}

		// Domain Eventに変換してマージ
		for _, model := range models {
			if evt, err := toDomain(model); err == nil {
				results = append(results, evt)
			} else {
				return []domain.Event{}, fmt.Errorf("failed to load events: %w", err)
			}
		}
	}

	// IDで重複除去（OR条件のため）
	results = domain.DedupeByID(results)

	return results, nil
}
