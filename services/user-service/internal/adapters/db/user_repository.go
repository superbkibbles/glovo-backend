package db

import (
	"glovo-backend/services/user-service/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhoneNumber(phoneNumber string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("phone_number = ?", phoneNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.User{}).Error
}

func (r *userRepository) List(limit, offset int) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}
