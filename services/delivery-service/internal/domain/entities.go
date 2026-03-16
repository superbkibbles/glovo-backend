package domain

import (
	"time"

	"github.com/google/uuid"
)

// AssignmentStatus represents delivery assignment status
type AssignmentStatus string

const (
	AssignmentStatusAssigned  AssignmentStatus = "assigned"
	AssignmentStatusPickedUp AssignmentStatus = "picked_up"
	AssignmentStatusDelivered AssignmentStatus = "delivered"
)

// DeliveryAssignment represents a delivery assignment
type DeliveryAssignment struct {
	ID          uuid.UUID        `gorm:"type:uuid;primary_key"`
	OrderID     uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex"`
	DriverID    uuid.UUID        `gorm:"type:uuid;not null"`
	DriverName  string
	Status      AssignmentStatus `gorm:"type:varchar(50);default:assigned"`
	DriverLat   float64
	DriverLng   float64
	AssignedAt  time.Time
	PickedUpAt  *time.Time
	DeliveredAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (DeliveryAssignment) TableName() string { return "delivery_assignments" }

// DeliveryLocationHistory stores tracking history for deliveries
type DeliveryLocationHistory struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	AssignmentID uuid.UUID `gorm:"type:uuid;not null;index"`
	DriverID     uuid.UUID `gorm:"type:uuid;not null"`
	Lat          float64   `gorm:"not null"`
	Lng          float64   `gorm:"not null"`
	RecordedAt   time.Time `gorm:"not null"`
}

func (DeliveryLocationHistory) TableName() string { return "delivery_location_history" }
