package domain

import (
	"time"

	"github.com/google/uuid"
)

// Restaurant represents a restaurant
type Restaurant struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Name         string    `gorm:"not null"`
	Description  string
	Logo         string
	CoverImage   string
	Address      string
	Lat          float64
	Lng          float64
	OpeningHours string
	Status       string    `gorm:"default:active"` // active, inactive
	OwnerID      uuid.UUID `gorm:"type:uuid"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Restaurant) TableName() string { return "restaurants" }

// Category represents a menu category
type Category struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Name         string    `gorm:"not null"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null"`
	SortOrder    int32     `gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Category) TableName() string { return "categories" }

// MenuItem represents a menu item
type MenuItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string    `gorm:"not null"`
	Description string
	Price       float64   `gorm:"not null"`
	Image       string
	CategoryID  uuid.UUID `gorm:"type:uuid;not null"`
	Available   bool      `gorm:"default:true"`
	Options     string    `gorm:"type:text"` // JSON
	SortOrder   int32     `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (MenuItem) TableName() string { return "menu_items" }
