package repository

import (
	"103-EmailService/pkg/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	Create(ctx context.Context, CreateAlertRequest *models.Alert) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Alert, error)
	Update(ctx context.Context, CreateAlertRequest *models.Alert) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context) ([]*models.Alert, error)
}

type AlertRepository struct {
	collection *mongo.Collection
}

func NewAlertRepository(db *mongo.Database, collectionName string) *AlertRepository {
	return &AlertRepository{
		collection: db.Collection(collectionName),
	}
}

func (alertRepo *AlertRepository) Create(ctx context.Context, alert *models.Alert) error {
	_, err := alertRepo.collection.InsertOne(ctx, alert)
	return err
}

func (alertRepo *AlertRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Alert, error) {
	var alert models.Alert
	err := alertRepo.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&alert)
	return &alert, err
}

func (alertRepo *AlertRepository) Update(ctx context.Context, alert *models.Alert) error {
	_, err := alertRepo.collection.UpdateOne(
		ctx,
		bson.M{"_id": alert.MigrationId},
		bson.M{"$set": alert},
		options.Update().SetUpsert(true),
	)
	return err
}

func (alertRepo *AlertRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := alertRepo.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (alertRepo *AlertRepository) List(ctx context.Context) ([]*models.Alert, error) {
	cursor, err := alertRepo.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var alerts []*models.Alert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}
