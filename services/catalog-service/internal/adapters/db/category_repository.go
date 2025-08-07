package db

import (
	"glovo-backend/services/catalog-service/internal/domain"

	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *domain.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetByID(id string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.Preload("Children").Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetAll() ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.Preload("Children").Order("sort_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetByParentID(parentID *string) ([]domain.Category, error) {
	var categories []domain.Category
	query := r.db.Preload("Children").Order("sort_order ASC, name ASC")

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(category *domain.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Category{}).Error
}
