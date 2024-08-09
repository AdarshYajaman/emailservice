package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository[T any] interface {
	Create(ctx context.Context, entity T) error
	GetByID(ctx context.Context, id primitive.ObjectID) (T, error)
	Update(ctx context.Context, entity T, id primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID) (int64, error)
	List(ctx context.Context, filter interface{}) ([]T, error)
}

type CustomRepository[T any] struct {
	collection *mongo.Collection
}

func NewCustomRepository[T any](db *mongo.Database, collectionName string) Repository[T] {
	return &CustomRepository[T]{
		collection: db.Collection(collectionName),
	}
}

func (repo *CustomRepository[T]) Create(ctx context.Context, entity T) error {
	_, err := repo.collection.InsertOne(ctx, entity)
	return err
}

func (repo *CustomRepository[T]) GetByID(ctx context.Context, id primitive.ObjectID) (T, error) {
	var entity T
	err := repo.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&entity)
	return entity, err
}

func (repo *CustomRepository[T]) Update(ctx context.Context, entity T, id primitive.ObjectID) error {
	_, err := repo.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": entity},
		options.Update().SetUpsert(false),
	)
	return err
}

func (repo *CustomRepository[T]) Delete(ctx context.Context, id primitive.ObjectID) (int64, error) {
	result, err := repo.collection.DeleteOne(ctx, bson.M{"_id": id})
	return result.DeletedCount, err
}

func (repo *CustomRepository[T]) List(ctx context.Context, filter interface{}) ([]T, error) {
	cursor, err := repo.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var entities []T
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}
