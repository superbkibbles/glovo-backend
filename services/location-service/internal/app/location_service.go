package app

import (
	"fmt"
	"math"
	"time"

	"glovo-backend/services/location-service/internal/domain"
)

type locationService struct {
	driverLocationRepo  domain.DriverLocationRepository
	locationHistoryRepo domain.LocationHistoryRepository
	deliveryRouteRepo   domain.DeliveryRouteRepository
	geofenceRepo        domain.GeofenceRepository
	mapsService         domain.MapsService
	notificationService domain.NotificationService
}

func NewLocationService(
	driverLocationRepo domain.DriverLocationRepository,
	locationHistoryRepo domain.LocationHistoryRepository,
	deliveryRouteRepo domain.DeliveryRouteRepository,
	geofenceRepo domain.GeofenceRepository,
	mapsService domain.MapsService,
	notificationService domain.NotificationService,
) domain.LocationService {
	return &locationService{
		driverLocationRepo:  driverLocationRepo,
		locationHistoryRepo: locationHistoryRepo,
		deliveryRouteRepo:   deliveryRouteRepo,
		geofenceRepo:        geofenceRepo,
		mapsService:         mapsService,
		notificationService: notificationService,
	}
}

// Driver location management
func (s *locationService) UpdateDriverLocation(driverID string, req domain.UpdateLocationRequest) (*domain.LocationResponse, error) {
	// Create GeoPoint
	location := domain.GeoPoint{
		Type:        "Point",
		Coordinates: []float64{req.Longitude, req.Latitude},
		Timestamp:   time.Now(),
	}

	// Determine status based on speed
	status := domain.StatusStopped
	if req.Speed > 1.0 { // Moving if speed > 1 km/h
		status = domain.StatusMoving
	}

	// Create or update driver location
	driverLocation := &domain.DriverLocation{
		DriverID:  driverID,
		Location:  location,
		Heading:   req.Heading,
		Speed:     req.Speed,
		Accuracy:  req.Accuracy,
		Altitude:  req.Altitude,
		Status:    status,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	// Check if this is the first location update
	existing, _ := s.driverLocationRepo.GetByDriverID(driverID)
	if existing != nil {
		driverLocation.CreatedAt = existing.CreatedAt
	}

	// Save location
	if err := s.driverLocationRepo.Upsert(driverLocation); err != nil {
		return nil, fmt.Errorf("failed to update driver location: %w", err)
	}

	// Check for geofence events
	go s.processGeofenceEvents(driverID, location)

	// Update active delivery route if exists
	go s.updateActiveRouteProgress(driverID, location)

	return &domain.LocationResponse{
		DriverID:  driverID,
		Location:  location,
		Heading:   req.Heading,
		Speed:     req.Speed,
		Status:    status,
		UpdatedAt: time.Now(),
	}, nil
}

func (s *locationService) GetDriverLocation(driverID string) (*domain.LocationResponse, error) {
	location, err := s.driverLocationRepo.GetByDriverID(driverID)
	if err != nil {
		return nil, err
	}

	return &domain.LocationResponse{
		DriverID:  location.DriverID,
		Location:  location.Location,
		Heading:   location.Heading,
		Speed:     location.Speed,
		Status:    location.Status,
		UpdatedAt: location.UpdatedAt,
	}, nil
}

func (s *locationService) GetNearbyDrivers(req domain.NearbyDriversRequest) ([]domain.NearbyDriver, error) {
	if req.Limit == 0 {
		req.Limit = 10 // Default limit
	}

	return s.driverLocationRepo.GetNearbyDrivers(req.Latitude, req.Longitude, req.Radius, req.Limit)
}

func (s *locationService) SetDriverStatus(driverID string, status domain.LocationStatus) error {
	return s.driverLocationRepo.UpdateStatus(driverID, status)
}

// Route management
func (s *locationService) CreateDeliveryRoute(req domain.RouteRequest) (*domain.DeliveryRoute, error) {
	// Calculate route using external maps service
	routeInfo, err := s.mapsService.CalculateRoute(req.PickupLocation, req.DropoffLocation)
	if err != nil {
		// Fallback to simple calculation
		routeInfo = &domain.RouteInfo{
			Distance:    s.calculateDistance(req.PickupLocation, req.DropoffLocation),
			Duration:    30, // Default 30 minutes
			RoutePoints: []domain.GeoPoint{req.PickupLocation, req.DropoffLocation},
		}
	}

	route := &domain.DeliveryRoute{
		OrderID:           req.OrderID,
		DriverID:          req.DriverID,
		PickupLocation:    req.PickupLocation,
		DropoffLocation:   req.DropoffLocation,
		RoutePoints:       routeInfo.RoutePoints,
		Status:            domain.RouteStatusActive,
		EstimatedDistance: routeInfo.Distance,
		EstimatedDuration: routeInfo.Duration,
		ActualDistance:    0,
		StartedAt:         time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.deliveryRouteRepo.Create(route); err != nil {
		return nil, fmt.Errorf("failed to create delivery route: %w", err)
	}

	// Start location tracking for this delivery
	s.StartLocationTracking(req.DriverID, &req.OrderID)

	return route, nil
}

func (s *locationService) GetDeliveryRoute(orderID string) (*domain.DeliveryRoute, error) {
	return s.deliveryRouteRepo.GetByOrderID(orderID)
}

func (s *locationService) UpdateRouteProgress(orderID string, currentLocation domain.GeoPoint) (*domain.DeliveryRoute, error) {
	route, err := s.deliveryRouteRepo.GetByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	// Update current location
	route.CurrentLocation = &currentLocation

	// Calculate actual distance traveled
	if route.CurrentLocation != nil {
		route.ActualDistance = s.calculateDistance(route.PickupLocation, currentLocation)
	}

	route.UpdatedAt = time.Now()

	if err := s.deliveryRouteRepo.Update(route); err != nil {
		return nil, err
	}

	return route, nil
}

func (s *locationService) CompleteRoute(orderID string) (*domain.DeliveryRoute, error) {
	if err := s.deliveryRouteRepo.Complete(orderID); err != nil {
		return nil, err
	}

	// Get the updated route and return it
	return s.deliveryRouteRepo.GetByOrderID(orderID)
}

func (s *locationService) GetActiveRoutes() ([]domain.DeliveryRoute, error) {
	return s.deliveryRouteRepo.GetActiveRoutes()
}

// Tracking for customers
func (s *locationService) GetOrderTrackingInfo(orderID string) (*domain.TrackingInfo, error) {
	route, err := s.deliveryRouteRepo.GetByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	trackingInfo := &domain.TrackingInfo{
		OrderID:  orderID,
		Distance: route.EstimatedDistance,
		Duration: route.EstimatedDuration,
		Status:   route.Status,
	}

	// Get current driver location
	if route.DriverID != "" {
		driverLocation, err := s.driverLocationRepo.GetByDriverID(route.DriverID)
		if err == nil {
			trackingInfo.DriverLocation = &driverLocation.Location

			// Calculate ETA if we have current location
			if route.CurrentLocation != nil {
				eta, err := s.mapsService.GetETA(*route.CurrentLocation, route.DropoffLocation)
				if err == nil {
					trackingInfo.EstimatedArrival = &eta.ArrivalTime
				}
			}
		}
	}

	return trackingInfo, nil
}

// Location history
func (s *locationService) GetLocationHistory(req domain.LocationHistoryRequest) ([]domain.LocationHistory, error) {
	return s.locationHistoryRepo.GetByDriverID(req.DriverID, req.StartTime, req.EndTime)
}

func (s *locationService) StartLocationTracking(driverID string, orderID *string) error {
	// Create new location history entry
	history := &domain.LocationHistory{
		DriverID:  driverID,
		OrderID:   orderID,
		Route:     []domain.GeoPoint{},
		Distance:  0,
		Duration:  0,
		StartTime: time.Now(),
		CreatedAt: time.Now(),
	}

	return s.locationHistoryRepo.Create(history)
}

func (s *locationService) StopLocationTracking(driverID string) error {
	// Find active tracking session and update end time
	histories, err := s.locationHistoryRepo.GetByDriverID(driverID, time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return err
	}

	// Find the most recent active session
	for _, history := range histories {
		if history.EndTime == nil {
			now := time.Now()
			history.EndTime = &now
			history.Duration = int64(now.Sub(history.StartTime).Seconds())

			return s.locationHistoryRepo.Update(&history)
		}
	}

	return nil
}

// Geofencing
func (s *locationService) CreateGeofence(geofence *domain.Geofence) (*domain.Geofence, error) {
	geofence.CreatedAt = time.Now()
	geofence.UpdatedAt = time.Now()

	if err := s.geofenceRepo.Create(geofence); err != nil {
		return nil, err
	}

	return geofence, nil
}

func (s *locationService) GetGeofences() ([]domain.Geofence, error) {
	return s.geofenceRepo.GetAll()
}

func (s *locationService) CheckGeofenceEvents(driverID string, location domain.GeoPoint) ([]domain.GeofenceEvent, error) {
	geofences, err := s.geofenceRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var events []domain.GeofenceEvent

	for _, geofence := range geofences {
		if !geofence.IsActive {
			continue
		}

		isInside := s.isPointInGeofence(location, &geofence)

		// For simplicity, we'll just create enter events when driver is inside
		// In a real implementation, you'd track previous state to determine enter/exit
		if isInside {
			event := domain.GeofenceEvent{
				DriverID:   driverID,
				GeofenceID: geofence.ID.Hex(),
				EventType:  domain.EventTypeEnter,
				Location:   location,
				Timestamp:  time.Now(),
			}
			events = append(events, event)
		}
	}

	// Send notifications for events
	for _, event := range events {
		go s.notificationService.SendGeofenceAlert(event)
	}

	return events, nil
}

// Analytics
func (s *locationService) GetDriverDistanceStats(driverID string, startTime, endTime time.Time) (*domain.DriverStats, error) {
	histories, err := s.locationHistoryRepo.GetByDriverID(driverID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	stats := &domain.DriverStats{
		DriverID: driverID,
	}

	for _, history := range histories {
		stats.TotalDistance += history.Distance
		stats.TotalDuration += history.Duration
		if history.OrderID != nil {
			stats.DeliveryCount++
		}
	}

	if stats.TotalDuration > 0 {
		stats.AverageSpeed = stats.TotalDistance / (float64(stats.TotalDuration) / 3600) // km/h
		stats.ActiveTime = stats.TotalDuration
	}

	return stats, nil
}

func (s *locationService) GetSystemLocationStats() (*domain.SystemStats, error) {
	// Get active routes for active deliveries count
	activeRoutes, err := s.deliveryRouteRepo.GetActiveRoutes()
	if err != nil {
		return nil, err
	}

	stats := &domain.SystemStats{
		ActiveDeliveries: len(activeRoutes),
	}

	// Calculate average delivery time from completed routes
	totalDuration := 0
	count := 0
	for _, route := range activeRoutes {
		if route.ActualDuration != nil {
			totalDuration += *route.ActualDuration
			count++
		}
	}

	if count > 0 {
		stats.AverageDeliveryTime = totalDuration / count
	}

	return stats, nil
}

// Helper functions
func (s *locationService) processGeofenceEvents(driverID string, location domain.GeoPoint) {
	events, err := s.CheckGeofenceEvents(driverID, location)
	if err != nil {
		return
	}

	// Events are already processed in CheckGeofenceEvents
	_ = events
}

func (s *locationService) updateActiveRouteProgress(driverID string, location domain.GeoPoint) {
	route, err := s.deliveryRouteRepo.GetByDriverID(driverID)
	if err != nil {
		return
	}

	if route.Status == domain.RouteStatusActive {
		s.UpdateRouteProgress(route.OrderID, location)
	}
}

func (s *locationService) calculateDistance(point1, point2 domain.GeoPoint) float64 {
	// Simple haversine distance calculation
	lat1 := point1.Coordinates[1]
	lon1 := point1.Coordinates[0]
	lat2 := point2.Coordinates[1]
	lon2 := point2.Coordinates[0]

	const R = 6371 // Earth's radius in kilometers

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (s *locationService) isPointInGeofence(point domain.GeoPoint, geofence *domain.Geofence) bool {
	// Simple circle geofence check
	if geofence.Geometry.Type == "Circle" {
		if geofence.Geometry.Radius == nil {
			return false
		}

		// Assume center coordinates are stored in Coordinates field
		center, ok := geofence.Geometry.Coordinates.([]float64)
		if !ok || len(center) != 2 {
			return false
		}

		centerPoint := domain.GeoPoint{
			Coordinates: center,
		}

		distance := s.calculateDistance(point, centerPoint)
		return distance <= (*geofence.Geometry.Radius / 1000) // Convert meters to km
	}

	// For polygon geofences, implement ray casting algorithm
	// This is a simplified version - in production, use a proper geospatial library
	return false
}
