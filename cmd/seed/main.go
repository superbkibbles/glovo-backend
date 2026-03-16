package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/rs/zerolog"
	"github.com/mendmzury/food-delivery/pkg/database"
	"github.com/mendmzury/food-delivery/pkg/logger"
	"github.com/mendmzury/food-delivery/pkg/middleware"
	"github.com/mendmzury/food-delivery/pkg/utils"
	"gorm.io/gorm"
)

// Seed Role/User structs (match auth-service schema)
type seedRole struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Icon        string
	Permissions seedStringArray `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (seedRole) TableName() string { return "roles" }

type seedUser struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Username     string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:text"`
	Salt         string    `gorm:"type:text"`
	Name         string
	Email        string `gorm:"index"`
	PhoneNumber  string `gorm:"index"`
	GoogleID     string `gorm:"index;column:google_id"`
	ProfilePhoto string
	RoleID       uuid.UUID `gorm:"type:uuid;not null"`
	UserType     string    `gorm:"type:varchar(50)"`
	IsSuperadmin bool      `gorm:"default:false"`
	Active       bool      `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (seedUser) TableName() string { return "users" }

type seedStringArray []string

func (s seedStringArray) Value() (driver.Value, error) { return json.Marshal(s) }
func (s *seedStringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for StringArray")
	}
	return json.Unmarshal(b, s)
}

// Restaurant entities for seed
type seedRestaurant struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string    `gorm:"not null"`
	Description string
	Logo        string
	CoverImage  string
	Address     string
	Lat         float64
	Lng         float64
	OpeningHours string
	Status      string `gorm:"default:active"` // active, inactive
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (seedRestaurant) TableName() string { return "restaurants" }

type seedCategory struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Name         string    `gorm:"not null"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null"`
	SortOrder    int32     `gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (seedCategory) TableName() string { return "categories" }

type seedMenuItem struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	Name         string    `gorm:"not null"`
	Description  string
	Price        float64   `gorm:"not null"`
	Image        string
	CategoryID   uuid.UUID `gorm:"type:uuid;not null"`
	Available    bool      `gorm:"default:true"`
	Options      string    `gorm:"type:text"` // JSON
	SortOrder    int32     `gorm:"default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (seedMenuItem) TableName() string { return "menu_items" }

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.Get()

	log.Info().Msg("Starting seed process...")

	ctx := context.Background()

	// Connect to Postgres
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Migrate auth tables
	if err := db.WithContext(ctx).AutoMigrate(&seedRole{}, &seedUser{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate auth tables")
	}
	log.Info().Msg("Migrated auth tables")

	// Migrate restaurant tables
	if err := db.WithContext(ctx).AutoMigrate(&seedRestaurant{}, &seedCategory{}, &seedMenuItem{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate restaurant tables")
	}
	log.Info().Msg("Migrated restaurant tables")

	// Seed superadmin if not exists
	seedAuth(db, cfg, log)

	// Seed demo restaurants
	seedRestaurants(db, cfg, log)

	// Connect to MongoDB and create indexes
	mongoDB, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to MongoDB (optional)")
	} else {
		defer mongoDB.Close(ctx)
		if err := mongoDB.CreateIndexes(ctx); err != nil {
			log.Warn().Err(err).Msg("Failed to create MongoDB indexes")
		} else {
			log.Info().Msg("Created MongoDB indexes")
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("  Food Delivery Seed Completed Successfully")
	fmt.Println("========================================")
	fmt.Printf("  Superadmin Username: %s\n", cfg.SuperadminUsername)
	fmt.Printf("  Superadmin Password: %s\n", cfg.SuperadminPassword)
	fmt.Println("========================================")
	fmt.Println("  Please change the password after first login!")
	fmt.Println("========================================\n")
}

func seedAuth(db *gorm.DB, cfg *config.Config, log *zerolog.Logger) {
	ctx := context.Background()
	var roleCount int64
	db.WithContext(ctx).Model(&seedRole{}).Count(&roleCount)
	if roleCount == 0 {
		adminRole := &seedRole{
			ID:          uuid.New(),
			Name:        "Admin",
			Icon:        "shield",
			Permissions: seedStringArray(middleware.AllPermissions()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := db.Create(adminRole).Error; err != nil {
			log.Fatal().Err(err).Msg("Failed to create Admin role")
		}
		log.Info().Str("role_id", adminRole.ID.String()).Msg("Created Admin role")
	}

	// Ensure Customer role exists for OTP signup
	var customerRole seedRole
	if err := db.WithContext(ctx).Where("name = ?", "Customer").First(&customerRole).Error; err != nil {
		customerRole = seedRole{
			ID:          uuid.New(),
			Name:        "Customer",
			Icon:        "user",
			Permissions: seedStringArray(middleware.CustomerPermissions()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := db.Create(&customerRole).Error; err != nil {
			log.Warn().Err(err).Msg("Failed to create Customer role")
		} else {
			log.Info().Str("role_id", customerRole.ID.String()).Msg("Created Customer role")
		}
	}

	var user seedUser
	if err := db.Where("is_superadmin = ?", true).First(&user).Error; err == nil {
		log.Info().Msg("Superadmin already exists, skipping")
		return
	}

	var adminRole seedRole
	if err := db.Where("name = ?", "Admin").First(&adminRole).Error; err != nil {
		log.Fatal().Err(err).Msg("Admin role not found")
	}

	pwHasher := utils.NewPasswordHasher(cfg.PasswordPepper)
	salt, _ := pwHasher.GenerateSalt()
	hash, _ := pwHasher.HashPassword(cfg.SuperadminPassword, salt)

	superadmin := &seedUser{
		ID:           uuid.New(),
		Username:     cfg.SuperadminUsername,
		PasswordHash: hash,
		Salt:         salt,
		Name:         "Superadmin",
		RoleID:       adminRole.ID,
		UserType:     "admin",
		IsSuperadmin: true,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(superadmin).Error; err != nil {
		log.Fatal().Err(err).Msg("Failed to create superadmin user")
	}
	log.Info().Str("username", superadmin.Username).Msg("Created Superadmin user")
}

func seedRestaurants(db *gorm.DB, cfg *config.Config, log *zerolog.Logger) {
	var count int64
	db.Model(&seedRestaurant{}).Count(&count)
	if count > 0 {
		log.Info().Msg("Restaurants already exist, skipping")
		return
	}

	// Get admin user for owner_id
	var adminUser seedUser
	if err := db.Where("is_superadmin = ?", true).First(&adminUser).Error; err != nil {
		log.Warn().Msg("No admin user for restaurant owner, using zero UUID")
	}

	r1 := &seedRestaurant{
		ID:           uuid.New(),
		Name:         "Pizza Palace",
		Description:  "Best pizza in town",
		Address:      "123 Main St",
		Lat:          36.1911,
		Lng:          44.0092,
		OpeningHours: "10:00-22:00",
		Status:       "active",
		OwnerID:      adminUser.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Create(r1)

	cat1 := &seedCategory{ID: uuid.New(), Name: "Pizza", RestaurantID: r1.ID, SortOrder: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(cat1)
	db.Create(&seedMenuItem{ID: uuid.New(), Name: "Margherita", Description: "Classic tomato and mozzarella", Price: 12.99, CategoryID: cat1.ID, Available: true, SortOrder: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	db.Create(&seedMenuItem{ID: uuid.New(), Name: "Pepperoni", Description: "Spicy pepperoni", Price: 14.99, CategoryID: cat1.ID, Available: true, SortOrder: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})

	r2 := &seedRestaurant{
		ID:           uuid.New(),
		Name:         "Burger Haven",
		Description:  "Gourmet burgers",
		Address:      "456 Oak Ave",
		Lat:          36.1920,
		Lng:          44.0100,
		OpeningHours: "11:00-23:00",
		Status:       "active",
		OwnerID:      adminUser.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Create(r2)

	cat2 := &seedCategory{ID: uuid.New(), Name: "Burgers", RestaurantID: r2.ID, SortOrder: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(cat2)
	db.Create(&seedMenuItem{ID: uuid.New(), Name: "Classic Burger", Description: "Beef patty with lettuce and tomato", Price: 9.99, CategoryID: cat2.ID, Available: true, SortOrder: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	db.Create(&seedMenuItem{ID: uuid.New(), Name: "Cheese Burger", Description: "Double cheese", Price: 11.99, CategoryID: cat2.ID, Available: true, SortOrder: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})

	log.Info().Msg("Seeded demo restaurants")
}
