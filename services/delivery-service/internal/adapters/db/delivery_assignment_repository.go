package db

import (
	"time"

	"glovo-backend/services/delivery-service/internal/domain"

	"gorm.io/gorm"
)

type deliveryAssignmentRepository struct {
	db *gorm.DB
}

func NewDeliveryAssignmentRepository(db *gorm.DB) domain.DeliveryAssignmentRepository {
	return &deliveryAssignmentRepository{db: db}
}

func (r *deliveryAssignmentRepository) Create(assignment *domain.DeliveryAssignment) error {
	return r.db.Create(assignment).Error
}

func (r *deliveryAssignmentRepository) GetByID(id string) (*domain.DeliveryAssignment, error) {
	var assignment domain.DeliveryAssignment
	err := r.db.Where("id = ?", id).First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *deliveryAssignmentRepository) GetByDeliveryID(deliveryID string) ([]domain.DeliveryAssignment, error) {
	var assignments []domain.DeliveryAssignment
	err := r.db.Where("delivery_id = ?", deliveryID).Find(&assignments).Error
	return assignments, err
}

func (r *deliveryAssignmentRepository) GetByDriverID(driverID string, limit, offset int) ([]domain.DeliveryAssignment, error) {
	var assignments []domain.DeliveryAssignment
	err := r.db.Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&assignments).Error
	return assignments, err
}

func (r *deliveryAssignmentRepository) GetPendingForDriver(driverID string) ([]domain.DeliveryAssignment, error) {
	var assignments []domain.DeliveryAssignment
	err := r.db.Where("driver_id = ? AND status = ?", driverID, domain.AssignmentPending).
		Find(&assignments).Error
	return assignments, err
}

func (r *deliveryAssignmentRepository) Update(assignment *domain.DeliveryAssignment) error {
	return r.db.Save(assignment).Error
}

func (r *deliveryAssignmentRepository) ExpirePendingAssignments() error {
	return r.db.Model(&domain.DeliveryAssignment{}).
		Where("status = ? AND expires_at < ?", domain.AssignmentPending, time.Now()).
		Update("status", domain.AssignmentExpired).Error
}
