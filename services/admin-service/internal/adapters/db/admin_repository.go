package db

import (
	"time"

	"glovo-backend/services/admin-service/internal/domain"

	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(admin *domain.Admin) error {
	return r.db.Create(admin).Error
}

func (r *adminRepository) GetByID(id string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Where("id = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByEmail(email string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) Update(admin *domain.Admin) error {
	return r.db.Save(admin).Error
}

func (r *adminRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Admin{}).Error
}

func (r *adminRepository) List(limit, offset int) ([]domain.Admin, error) {
	var admins []domain.Admin
	err := r.db.Limit(limit).Offset(offset).Find(&admins).Error
	return admins, err
}

func (r *adminRepository) UpdateLastLogin(id string) error {
	now := time.Now()
	return r.db.Model(&domain.Admin{}).Where("id = ?", id).Update("last_login_at", &now).Error
}
