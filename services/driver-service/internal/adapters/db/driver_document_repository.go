package db

import (
	"glovo-backend/services/driver-service/internal/domain"

	"gorm.io/gorm"
)

type driverDocumentRepository struct {
	db *gorm.DB
}

func NewDriverDocumentRepository(db *gorm.DB) domain.DriverDocumentRepository {
	return &driverDocumentRepository{db: db}
}

func (r *driverDocumentRepository) Create(document *domain.DriverDocument) error {
	return r.db.Create(document).Error
}

func (r *driverDocumentRepository) GetByID(id string) (*domain.DriverDocument, error) {
	var document domain.DriverDocument
	err := r.db.Where("id = ?", id).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *driverDocumentRepository) GetByDriverID(driverID string) ([]domain.DriverDocument, error) {
	var documents []domain.DriverDocument
	err := r.db.Where("driver_id = ?", driverID).
		Order("uploaded_at DESC").
		Find(&documents).Error
	return documents, err
}

func (r *driverDocumentRepository) Update(document *domain.DriverDocument) error {
	return r.db.Save(document).Error
}

func (r *driverDocumentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.DriverDocument{}).Error
}

func (r *driverDocumentRepository) GetByStatusAndType(status domain.DocumentStatus, docType domain.DocumentType) ([]domain.DriverDocument, error) {
	var documents []domain.DriverDocument
	err := r.db.Where("status = ? AND type = ?", status, docType).
		Order("uploaded_at DESC").
		Find(&documents).Error
	return documents, err
}
