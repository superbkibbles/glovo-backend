package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/domain"
	"gorm.io/gorm"
)

type AssignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) Create(ctx context.Context, a *domain.DeliveryAssignment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *AssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DeliveryAssignment, error) {
	var a domain.DeliveryAssignment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AssignmentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.DeliveryAssignment, error) {
	var a domain.DeliveryAssignment
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AssignmentRepository) List(ctx context.Context, driverID *uuid.UUID, status domain.AssignmentStatus, page, pageSize int64) ([]*domain.DeliveryAssignment, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.DeliveryAssignment{})

	if driverID != nil {
		query = query.Where("driver_id = ?", *driverID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var assignments []*domain.DeliveryAssignment
	err := query.Order("assigned_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&assignments).Error
	return assignments, total, err
}

func (r *AssignmentRepository) Update(ctx context.Context, a *domain.DeliveryAssignment) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *AssignmentRepository) GetActiveByDriverID(ctx context.Context, driverID uuid.UUID) (*domain.DeliveryAssignment, error) {
	var a domain.DeliveryAssignment
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND status IN ?", driverID, []domain.AssignmentStatus{
			domain.AssignmentStatusAssigned,
			domain.AssignmentStatusPickedUp,
		}).
		Order("assigned_at DESC").
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}
