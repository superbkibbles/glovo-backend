package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/domain"
	"gorm.io/gorm"
)

type TrackingRepository struct {
	db *gorm.DB
}

func NewTrackingRepository(db *gorm.DB) *TrackingRepository {
	return &TrackingRepository{db: db}
}

func (r *TrackingRepository) Create(ctx context.Context, h *domain.DeliveryLocationHistory) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *TrackingRepository) ListByAssignmentID(ctx context.Context, assignmentID uuid.UUID) ([]*domain.DeliveryLocationHistory, error) {
	var points []*domain.DeliveryLocationHistory
	err := r.db.WithContext(ctx).Where("assignment_id = ?", assignmentID).
		Order("recorded_at ASC").
		Find(&points).Error
	return points, err
}
