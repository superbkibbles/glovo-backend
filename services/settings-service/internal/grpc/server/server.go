package server

import (
	"context"
	"time"

	"github.com/mendmzury/food-delivery/pkg/database"
	settingspb "github.com/mendmzury/food-delivery/proto/settings"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const commissionKey = "app_commission"

type SettingsServer struct {
	settingspb.UnimplementedSettingsServiceServer
	mongo *database.MongoDB
}

func NewSettingsServer(mongo *database.MongoDB) *SettingsServer {
	return &SettingsServer{mongo: mongo}
}

func (s *SettingsServer) GetCommission(ctx context.Context, req *settingspb.GetCommissionRequest) (*settingspb.GetCommissionResponse, error) {
	coll := s.mongo.Collection(database.CollectionSettings)
	var doc struct {
		Key   string  `bson:"key"`
		Value float64 `bson:"value"`
	}
	err := coll.FindOne(ctx, bson.M{"key": commissionKey}).Decode(&doc)
	if err != nil {
		return &settingspb.GetCommissionResponse{CommissionPercent: 0}, nil
	}
	return &settingspb.GetCommissionResponse{CommissionPercent: doc.Value}, nil
}

func (s *SettingsServer) UpdateCommission(ctx context.Context, req *settingspb.UpdateCommissionRequest) (*settingspb.GetCommissionResponse, error) {
	if req.CommissionPercent < 0 || req.CommissionPercent > 100 {
		return nil, status.Error(codes.InvalidArgument, "commission must be between 0 and 100")
	}
	coll := s.mongo.Collection(database.CollectionSettings)
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(ctx,
		bson.M{"key": commissionKey},
		bson.M{"$set": bson.M{"key": commissionKey, "value": req.CommissionPercent}},
		opts,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &settingspb.GetCommissionResponse{CommissionPercent: req.CommissionPercent}, nil
}

func (s *SettingsServer) ListOperatingAreas(ctx context.Context, req *settingspb.ListOperatingAreasRequest) (*settingspb.ListOperatingAreasResponse, error) {
	coll := s.mongo.Collection(database.CollectionOperatingAreas)
	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer cur.Close(ctx)

	var areas []*settingspb.OperatingArea
	for cur.Next(ctx) {
		var doc operatingAreaDoc
		if err := cur.Decode(&doc); err != nil {
			continue
		}
		areas = append(areas, docToProto(&doc))
	}
	return &settingspb.ListOperatingAreasResponse{Areas: areas}, nil
}

func (s *SettingsServer) CreateOperatingArea(ctx context.Context, req *settingspb.CreateOperatingAreaRequest) (*settingspb.OperatingArea, error) {
	if req.Name == "" || len(req.Coordinates) < 3 {
		return nil, status.Error(codes.InvalidArgument, "name and at least 3 coordinates required")
	}
	coll := s.mongo.Collection(database.CollectionOperatingAreas)
	now := primitive.NewDateTimeFromTime(time.Now())
	doc := &operatingAreaDoc{
		Name:        req.Name,
		Coordinates: coordsToDoc(req.Coordinates),
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	res, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	doc.ID = res.InsertedID.(primitive.ObjectID)
	return docToProto(doc), nil
}

func (s *SettingsServer) UpdateOperatingArea(ctx context.Context, req *settingspb.UpdateOperatingAreaRequest) (*settingspb.OperatingArea, error) {
	oid, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	coll := s.mongo.Collection(database.CollectionOperatingAreas)
	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if len(req.Coordinates) >= 3 {
		update["coordinates"] = coordsToDoc(req.Coordinates)
	}
	update["active"] = req.Active
	update["updated_at"] = primitive.NewDateTimeFromTime(time.Now())
	_, err = coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var doc operatingAreaDoc
	if err := coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		return nil, status.Error(codes.NotFound, "operating area not found")
	}
	return docToProto(&doc), nil
}

func (s *SettingsServer) DeleteOperatingArea(ctx context.Context, req *settingspb.DeleteOperatingAreaRequest) (*settingspb.OperatingArea, error) {
	oid, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	coll := s.mongo.Collection(database.CollectionOperatingAreas)
	var doc operatingAreaDoc
	if err := coll.FindOneAndDelete(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		return nil, status.Error(codes.NotFound, "operating area not found")
	}
	return docToProto(&doc), nil
}

type operatingAreaDoc struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Coordinates []coordDoc         `bson:"coordinates"`
	Active      bool               `bson:"active"`
	CreatedAt   primitive.DateTime `bson:"created_at,omitempty"`
	UpdatedAt   primitive.DateTime `bson:"updated_at,omitempty"`
}

type coordDoc struct {
	Lng float64 `bson:"lng"`
	Lat float64 `bson:"lat"`
}

func coordsToDoc(coords []*settingspb.Coordinate) []coordDoc {
	out := make([]coordDoc, len(coords))
	for i, c := range coords {
		out[i] = coordDoc{Lng: c.Lng, Lat: c.Lat}
	}
	return out
}

func docToProto(doc *operatingAreaDoc) *settingspb.OperatingArea {
	coords := make([]*settingspb.Coordinate, len(doc.Coordinates))
	for i, c := range doc.Coordinates {
		coords[i] = &settingspb.Coordinate{Lng: c.Lng, Lat: c.Lat}
	}
	p := &settingspb.OperatingArea{
		Id:          doc.ID.Hex(),
		Name:        doc.Name,
		Coordinates: coords,
		Active:      doc.Active,
	}
	if doc.CreatedAt != 0 {
		p.CreatedAt = timestamppb.New(doc.CreatedAt.Time())
	}
	if doc.UpdatedAt != 0 {
		p.UpdatedAt = timestamppb.New(doc.UpdatedAt.Time())
	}
	return p
}
