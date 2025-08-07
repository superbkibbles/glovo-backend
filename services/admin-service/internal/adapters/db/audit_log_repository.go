package db

import (
	"glovo-backend/services/admin-service/internal/domain"

	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *domain.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) GetByAdminID(adminID string, limit, offset int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.Where("admin_id = ?", adminID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (r *auditLogRepository) GetByResource(resource string, limit, offset int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.Where("resource = ?", resource).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (r *auditLogRepository) List(limit, offset int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
