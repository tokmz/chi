package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository 仓储接口
type Repository struct {
	client     *Client
	collection *mongo.Collection
	logger     Logger
	slowLogger *SlowQueryLogger
}

// NewRepository 创建新的仓储实例
func NewRepository(client *Client, database, collection string) *Repository {
	coll := client.Collection(collection, database)
	slowLogger := NewSlowQueryLogger(
		client.GetLogger(),
		client.GetConfig().Log.SlowQueryThreshold,
		client.GetConfig().Log.SlowQuery,
	)

	return &Repository{
		client:     client,
		collection: coll,
		logger:     client.GetLogger(),
		slowLogger: slowLogger,
	}
}

// Collection 获取集合实例
func (r *Repository) Collection() *mongo.Collection {
	return r.collection
}

// Client 获取客户端实例
func (r *Repository) Client() *Client {
	return r.client
}

// InsertOne 插入单个文档
func (r *Repository) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("InsertOne", duration, r.collection.Name(), document)
	}()

	result, err := r.collection.InsertOne(ctx, document, opts...)
	if err != nil {
		r.logger.Error("Failed to insert document", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	r.logger.Debug("Document inserted successfully", "collection", r.collection.Name(), "id", result.InsertedID)
	return result, nil
}

// InsertMany 插入多个文档
func (r *Repository) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("InsertMany", duration, r.collection.Name(), fmt.Sprintf("count: %d", len(documents)))
	}()

	result, err := r.collection.InsertMany(ctx, documents, opts...)
	if err != nil {
		r.logger.Error("Failed to insert documents", "error", err, "collection", r.collection.Name(), "count", len(documents))
		return nil, fmt.Errorf("failed to insert documents: %w", err)
	}

	r.logger.Debug("Documents inserted successfully", "collection", r.collection.Name(), "count", len(result.InsertedIDs))
	return result, nil
}

// FindOne 查找单个文档
func (r *Repository) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("FindOne", duration, r.collection.Name(), filter)
	}()

	return r.collection.FindOne(ctx, filter, opts...)
}

// FindOneByID 根据ID查找单个文档
func (r *Repository) FindOneByID(ctx context.Context, id interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	var objectID primitive.ObjectID
	var err error

	switch v := id.(type) {
	case string:
		objectID, err = primitive.ObjectIDFromHex(v)
		if err != nil {
			r.logger.Error("Invalid ObjectID", "error", err, "id", v)
			return &mongo.SingleResult{}
		}
	case primitive.ObjectID:
		objectID = v
	default:
		r.logger.Error("Unsupported ID type", "type", fmt.Sprintf("%T", id))
		return &mongo.SingleResult{}
	}

	filter := bson.M{"_id": objectID}
	return r.FindOne(ctx, filter, opts...)
}

// Find 查找多个文档
func (r *Repository) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("Find", duration, r.collection.Name(), filter)
	}()

	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		r.logger.Error("Failed to find documents", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}

	return cursor, nil
}

// FindAll 查找所有匹配的文档并解码到结果中
func (r *Repository) FindAll(ctx context.Context, filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	cursor, err := r.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		r.logger.Error("Failed to decode documents", "error", err, "collection", r.collection.Name())
		return fmt.Errorf("failed to decode documents: %w", err)
	}

	return nil
}

// UpdateOne 更新单个文档
func (r *Repository) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("UpdateOne", duration, r.collection.Name(), filter)
	}()

	result, err := r.collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		r.logger.Error("Failed to update document", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	r.logger.Debug("Document updated successfully", "collection", r.collection.Name(), "matched", result.MatchedCount, "modified", result.ModifiedCount)
	return result, nil
}

// UpdateOneByID 根据ID更新单个文档
func (r *Repository) UpdateOneByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var objectID primitive.ObjectID
	var err error

	switch v := id.(type) {
	case string:
		objectID, err = primitive.ObjectIDFromHex(v)
		if err != nil {
			r.logger.Error("Invalid ObjectID", "error", err, "id", v)
			return nil, ErrInvalidObjectID
		}
	case primitive.ObjectID:
		objectID = v
	default:
		r.logger.Error("Unsupported ID type", "type", fmt.Sprintf("%T", id))
		return nil, ErrInvalidObjectID
	}

	filter := bson.M{"_id": objectID}
	return r.UpdateOne(ctx, filter, update, opts...)
}

// UpdateMany 更新多个文档
func (r *Repository) UpdateMany(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("UpdateMany", duration, r.collection.Name(), filter)
	}()

	result, err := r.collection.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		r.logger.Error("Failed to update documents", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to update documents: %w", err)
	}

	r.logger.Debug("Documents updated successfully", "collection", r.collection.Name(), "matched", result.MatchedCount, "modified", result.ModifiedCount)
	return result, nil
}

// ReplaceOne 替换单个文档
func (r *Repository) ReplaceOne(ctx context.Context, filter, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("ReplaceOne", duration, r.collection.Name(), filter)
	}()

	result, err := r.collection.ReplaceOne(ctx, filter, replacement, opts...)
	if err != nil {
		r.logger.Error("Failed to replace document", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to replace document: %w", err)
	}

	r.logger.Debug("Document replaced successfully", "collection", r.collection.Name(), "matched", result.MatchedCount, "modified", result.ModifiedCount)
	return result, nil
}

// DeleteOne 删除单个文档
func (r *Repository) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("DeleteOne", duration, r.collection.Name(), filter)
	}()

	result, err := r.collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		r.logger.Error("Failed to delete document", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	r.logger.Debug("Document deleted successfully", "collection", r.collection.Name(), "deleted", result.DeletedCount)
	return result, nil
}

// DeleteOneByID 根据ID删除单个文档
func (r *Repository) DeleteOneByID(ctx context.Context, id interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	var objectID primitive.ObjectID
	var err error

	switch v := id.(type) {
	case string:
		objectID, err = primitive.ObjectIDFromHex(v)
		if err != nil {
			r.logger.Error("Invalid ObjectID", "error", err, "id", v)
			return nil, ErrInvalidObjectID
		}
	case primitive.ObjectID:
		objectID = v
	default:
		r.logger.Error("Unsupported ID type", "type", fmt.Sprintf("%T", id))
		return nil, ErrInvalidObjectID
	}

	filter := bson.M{"_id": objectID}
	return r.DeleteOne(ctx, filter, opts...)
}

// DeleteMany 删除多个文档
func (r *Repository) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("DeleteMany", duration, r.collection.Name(), filter)
	}()

	result, err := r.collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		r.logger.Error("Failed to delete documents", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to delete documents: %w", err)
	}

	r.logger.Debug("Documents deleted successfully", "collection", r.collection.Name(), "deleted", result.DeletedCount)
	return result, nil
}

// CountDocuments 统计文档数量
func (r *Repository) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("CountDocuments", duration, r.collection.Name(), filter)
	}()

	count, err := r.collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		r.logger.Error("Failed to count documents", "error", err, "collection", r.collection.Name())
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// EstimatedDocumentCount 估算文档数量
func (r *Repository) EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("EstimatedDocumentCount", duration, r.collection.Name(), nil)
	}()

	count, err := r.collection.EstimatedDocumentCount(ctx, opts...)
	if err != nil {
		r.logger.Error("Failed to estimate document count", "error", err, "collection", r.collection.Name())
		return 0, fmt.Errorf("failed to estimate document count: %w", err)
	}

	return count, nil
}

// Distinct 获取不重复的字段值
func (r *Repository) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("Distinct", duration, r.collection.Name(), filter)
	}()

	result, err := r.collection.Distinct(ctx, fieldName, filter, opts...)
	if err != nil {
		r.logger.Error("Failed to get distinct values", "error", err, "collection", r.collection.Name(), "field", fieldName)
		return nil, fmt.Errorf("failed to get distinct values: %w", err)
	}

	return result, nil
}

// Aggregate 聚合查询
func (r *Repository) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("Aggregate", duration, r.collection.Name(), pipeline)
	}()

	cursor, err := r.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		r.logger.Error("Failed to aggregate", "error", err, "collection", r.collection.Name())
		return nil, fmt.Errorf("failed to aggregate: %w", err)
	}

	return cursor, nil
}

// AggregateAll 聚合查询并解码到结果中
func (r *Repository) AggregateAll(ctx context.Context, pipeline interface{}, results interface{}, opts ...*options.AggregateOptions) error {
	cursor, err := r.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		r.logger.Error("Failed to decode aggregate results", "error", err, "collection", r.collection.Name())
		return fmt.Errorf("failed to decode aggregate results: %w", err)
	}

	return nil
}

// BulkWrite 批量写操作
func (r *Repository) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.slowLogger.LogSlowQuery("BulkWrite", duration, r.collection.Name(), fmt.Sprintf("operations: %d", len(models)))
	}()

	result, err := r.collection.BulkWrite(ctx, models, opts...)
	if err != nil {
		r.logger.Error("Failed to execute bulk write", "error", err, "collection", r.collection.Name(), "operations", len(models))
		return nil, fmt.Errorf("failed to execute bulk write: %w", err)
	}

	r.logger.Debug("Bulk write completed successfully",
		"collection", r.collection.Name(),
		"inserted", result.InsertedCount,
		"matched", result.MatchedCount,
		"modified", result.ModifiedCount,
		"deleted", result.DeletedCount,
		"upserted", result.UpsertedCount,
	)

	return result, nil
}