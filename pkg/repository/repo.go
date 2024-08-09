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
	Create(ctx context.Context, alert *models.Alert) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Alert, error)
	Update(ctx context.Context, alert *models.Alert) error
	Delete(ctx context.Context, id primitive.ObjectID) (int64, error)
	List(ctx context.Context, filter interface{}) ([]*models.Alert, error)
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
		bson.M{"_id": alert.IndexId},
		bson.M{"$set": alert},
		options.Update().SetUpsert(false),
	)
	return err
}

func (alertRepo *AlertRepository) Delete(ctx context.Context, id primitive.ObjectID) (int64, error) {
	result, err := alertRepo.collection.DeleteOne(ctx, bson.M{"_id": id})
	return result.DeletedCount, err
}

func (alertRepo *AlertRepository) List(ctx context.Context, filter interface{}) ([]*models.Alert, error) {
	cursor, err := alertRepo.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var alerts []*models.Alert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

type Jrepository interface {
	Create(ctx context.Context, job *models.Job) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	Delete(ctx context.Context, id primitive.ObjectID) (int64, error)
	List(ctx context.Context, filter interface{}) ([]*models.Job, error)
}

type JobRepository struct {
	collection *mongo.Collection
}

func NewJobRepository(db *mongo.Database, collectionName string) *JobRepository {
	return &JobRepository{
		collection: db.Collection(collectionName),
	}
}

func (jobRepo *JobRepository) Create(ctx context.Context, job *models.Job) error {
	_, err := jobRepo.collection.InsertOne(ctx, job)
	return err
}

func (jobRepo *JobRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Job, error) {
	var job models.Job
	err := jobRepo.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	return &job, err
}

func (jobRepo *JobRepository) Update(ctx context.Context, job *models.Job) error {
	_, err := jobRepo.collection.UpdateOne(
		ctx,
		bson.M{"_id": job.IndexId},
		bson.M{"$set": job},
		options.Update().SetUpsert(false),
	)
	return err
}

func (jobRepo *JobRepository) Delete(ctx context.Context, id primitive.ObjectID) (int64, error) {
	result, err := jobRepo.collection.DeleteOne(ctx, bson.M{"_id": id})
	return result.DeletedCount, err
}

func (jobRepo *JobRepository) List(ctx context.Context, filter interface{}) ([]*models.Job, error) {
	cursor, err := jobRepo.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var jobs []*models.Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}
