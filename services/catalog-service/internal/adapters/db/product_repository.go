package db

import (
	"strings"

	"glovo-backend/services/catalog-service/internal/domain"

	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *domain.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetByID(id string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.Preload("Options.Options").Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetByStoreID(storeID string, limit, offset int) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Preload("Options.Options").
		Where("store_id = ?", storeID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetByCategoryID(categoryID string, limit, offset int) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Preload("Options.Options").
		Where("category_id = ?", categoryID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error
	return products, err
}

func (r *productRepository) Update(product *domain.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id string) error {
	// Delete associated options and choices first
	if err := r.db.Where("product_id = ?", id).Delete(&domain.ProductOption{}).Error; err != nil {
		return err
	}

	// Delete the product
	return r.db.Where("id = ?", id).Delete(&domain.Product{}).Error
}

func (r *productRepository) Search(query string, storeID string, limit, offset int) ([]domain.Product, error) {
	dbQuery := r.db.Model(&domain.Product{}).Preload("Options.Options")

	// Apply search filters
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR EXISTS (SELECT 1 FROM unnest(tags) AS tag WHERE LOWER(tag) LIKE ?)",
			searchTerm, searchTerm, searchTerm,
		)
	}

	if storeID != "" {
		dbQuery = dbQuery.Where("store_id = ?", storeID)
	}

	// Only show available products in search results
	dbQuery = dbQuery.Where("status = ?", domain.ProductStatusAvailable)

	var products []domain.Product
	err := dbQuery.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

	return products, err
}
