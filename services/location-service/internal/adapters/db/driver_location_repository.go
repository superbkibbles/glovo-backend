package db

import (
	"context"
	"time"

	"glovo-backend/services/location-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type driverLocationRepository struct {
	collection *mongo.Collection
}

func NewDriverLocationRepository(db *mongo.Database) domain.DriverLocationRepository {
	collection := db.Collection("driver_locations")

	// Create geospatial index for location queries
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "location", Value: "2dsphere"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection.Indexes().CreateOne(ctx, indexModel)

	// Create index on driver_id for faster queries
	driverIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "driver_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	collection.Indexes().CreateOne(ctx, driverIndexModel)

	return &driverLocationRepository{collection: collection}
}

func (r *driverLocationRepository) Upsert(location *domain.DriverLocation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"driver_id": location.DriverID}
	update := bson.M{
		"$set": bson.M{
			"driver_id":  location.DriverID,
			"location":   location.Location,
			"heading":    location.Heading,
			"speed":      location.Speed,
			"accuracy":   location.Accuracy,
			"altitude":   location.Altitude,
			"status":     location.Status,
			"updated_at": location.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"created_at": location.CreatedAt,
		},
	}

	options := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, options)
	return err
}

func (r *driverLocationRepository) GetByDriverID(driverID string) (*domain.DriverLocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var location domain.DriverLocation
	filter := bson.M{"driver_id": driverID}

	err := r.collection.FindOne(ctx, filter).Decode(&location)
	if err != nil {
		return nil, err
	}

	return &location, nil
}

func (r *driverLocationRepository) GetNearbyDrivers(latitude, longitude, radius float64, limit int) ([]domain.NearbyDriver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert radius from kilometers to meters for MongoDB
	radiusMeters := radius * 1000

	pipeline := []bson.M{
		{
			"$geoNear": bson.M{
				"near": bson.M{
					"type":        "Point",
					"coordinates": []float64{longitude, latitude},
				},
				"distanceField": "distance",
				"maxDistance":   radiusMeters,
				"spherical":     true,
			},
		},
		{
			"$match": bson.M{
				"status": bson.M{"$in": []string{string(domain.StatusOnline), string(domain.StatusStopped)}},
			},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		DriverID string                `bson:"driver_id"`
		Location domain.GeoPoint       `bson:"location"`
		Status   domain.LocationStatus `bson:"status"`
		Distance float64               `bson:"distance"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	nearbyDrivers := make([]domain.NearbyDriver, len(results))
	for i, result := range results {
		nearbyDrivers[i] = domain.NearbyDriver{
			DriverID: result.DriverID,
			Location: result.Location,
			Distance: result.Distance / 1000, // Convert back to kilometers
			Status:   result.Status,
		}
	}

	return nearbyDrivers, nil
}

func (r *driverLocationRepository) GetDriversInGeofence(geofence *domain.Geofence) ([]domain.DriverLocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var filter bson.M

	if geofence.Geometry.Type == "Circle" {
		if geofence.Geometry.Radius == nil {
			return nil, nil
		}

		center, ok := geofence.Geometry.Coordinates.([]float64)
		if !ok || len(center) != 2 {
			return nil, nil
		}

		filter = bson.M{
			"location": bson.M{
				"$geoWithin": bson.M{
					"$centerSphere": []interface{}{
						center,
						*geofence.Geometry.Radius / 6378100, // Convert meters to radians
					},
				},
			},
		}
	} else if geofence.Geometry.Type == "Polygon" {
		filter = bson.M{
			"location": bson.M{
				"$geoWithin": bson.M{
					"$geometry": geofence.Geometry,
				},
			},
		}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []domain.DriverLocation
	if err := cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}

	return drivers, nil
}

func (r *driverLocationRepository) UpdateStatus(driverID string, status domain.LocationStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"driver_id": driverID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *driverLocationRepository) Delete(driverID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"driver_id": driverID}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}
