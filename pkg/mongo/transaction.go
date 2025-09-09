package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// TransactionManager 事务管理器
type TransactionManager struct {
	client *Client
	logger Logger
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(client *Client) *TransactionManager {
	return &TransactionManager{
		client: client,
		logger: client.GetLogger(),
	}
}

// TransactionOptions 事务选项
type TransactionOptions struct {
	ReadConcern    *readconcern.ReadConcern
	WriteConcern   *writeconcern.WriteConcern
	ReadPreference *readpref.ReadPref
	MaxCommitTime  *time.Duration
}

// DefaultTransactionOptions 默认事务选项
func DefaultTransactionOptions() *TransactionOptions {
	maxCommitTime := 30 * time.Second
	return &TransactionOptions{
		ReadConcern:   readconcern.Snapshot(),
		WriteConcern:  writeconcern.Majority(),
		MaxCommitTime: &maxCommitTime,
	}
}

// ToSessionOptions 转换为会话选项
func (opts *TransactionOptions) ToSessionOptions() *options.SessionOptions {
	sessionOpts := options.Session()
	if opts.ReadConcern != nil {
		sessionOpts.SetDefaultReadConcern(opts.ReadConcern)
	}
	if opts.WriteConcern != nil {
		sessionOpts.SetDefaultWriteConcern(opts.WriteConcern)
	}
	if opts.ReadPreference != nil {
		sessionOpts.SetDefaultReadPreference(opts.ReadPreference)
	}
	if opts.MaxCommitTime != nil {
		sessionOpts.SetDefaultMaxCommitTime(opts.MaxCommitTime)
	}
	return sessionOpts
}

// TransactionFunc 事务函数类型
type TransactionFunc func(ctx mongo.SessionContext) (interface{}, error)

// WithTransaction 执行事务
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn TransactionFunc, opts ...*TransactionOptions) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		tm.logger.Debug("Transaction completed", "duration", duration.String())
	}()

	// 获取事务选项
	var transactionOpts *TransactionOptions
	if len(opts) > 0 && opts[0] != nil {
		transactionOpts = opts[0]
	} else {
		transactionOpts = DefaultTransactionOptions()
	}

	// 创建会话
	sessionOpts := transactionOpts.ToSessionOptions()
	session, err := tm.client.Client().StartSession(sessionOpts)
	if err != nil {
		tm.logger.Error("Failed to start session", "error", err)
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 执行事务
	tm.logger.Debug("Starting transaction")
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		_, err = session.WithTransaction(sc, fn)
		return err
	})

	if err != nil {
		tm.logger.Error("Transaction failed", "error", err)
		return fmt.Errorf("transaction failed: %w", err)
	}

	tm.logger.Debug("Transaction completed successfully")
	return nil
}

// WithSession 使用会话执行操作
func (tm *TransactionManager) WithSession(ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.SessionOptions) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		tm.logger.Debug("Session operation completed", "duration", duration.String())
	}()

	// 获取会话选项
	var sessionOpts *options.SessionOptions
	if len(opts) > 0 && opts[0] != nil {
		sessionOpts = opts[0]
	} else {
		sessionOpts = options.Session()
	}

	// 创建会话
	session, err := tm.client.Client().StartSession(sessionOpts)
	if err != nil {
		tm.logger.Error("Failed to start session", "error", err)
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 执行操作
	tm.logger.Debug("Starting session operation")
	err = mongo.WithSession(ctx, session, fn)
	if err != nil {
		tm.logger.Error("Session operation failed", "error", err)
		return fmt.Errorf("session operation failed: %w", err)
	}

	tm.logger.Debug("Session operation completed successfully")
	return nil
}

// TransactionRepository 事务仓储
type TransactionRepository struct {
	*Repository
	session mongo.SessionContext
}

// NewTransactionRepository 创建事务仓储
func NewTransactionRepository(repo *Repository, session mongo.SessionContext) *TransactionRepository {
	return &TransactionRepository{
		Repository: repo,
		session:    session,
	}
}

// InsertOne 在事务中插入单个文档
func (tr *TransactionRepository) InsertOne(document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return tr.Repository.InsertOne(tr.session, document, opts...)
}

// InsertMany 在事务中插入多个文档
func (tr *TransactionRepository) InsertMany(documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return tr.Repository.InsertMany(tr.session, documents, opts...)
}

// FindOne 在事务中查找单个文档
func (tr *TransactionRepository) FindOne(filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return tr.Repository.FindOne(tr.session, filter, opts...)
}

// FindOneByID 在事务中根据ID查找单个文档
func (tr *TransactionRepository) FindOneByID(id interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return tr.Repository.FindOneByID(tr.session, id, opts...)
}

// Find 在事务中查找多个文档
func (tr *TransactionRepository) Find(filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return tr.Repository.Find(tr.session, filter, opts...)
}

// FindAll 在事务中查找所有匹配的文档
func (tr *TransactionRepository) FindAll(filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	return tr.Repository.FindAll(tr.session, filter, results, opts...)
}

// UpdateOne 在事务中更新单个文档
func (tr *TransactionRepository) UpdateOne(filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return tr.Repository.UpdateOne(tr.session, filter, update, opts...)
}

// UpdateOneByID 在事务中根据ID更新单个文档
func (tr *TransactionRepository) UpdateOneByID(id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return tr.Repository.UpdateOneByID(tr.session, id, update, opts...)
}

// UpdateMany 在事务中更新多个文档
func (tr *TransactionRepository) UpdateMany(filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return tr.Repository.UpdateMany(tr.session, filter, update, opts...)
}

// ReplaceOne 在事务中替换单个文档
func (tr *TransactionRepository) ReplaceOne(filter, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	return tr.Repository.ReplaceOne(tr.session, filter, replacement, opts...)
}

// DeleteOne 在事务中删除单个文档
func (tr *TransactionRepository) DeleteOne(filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return tr.Repository.DeleteOne(tr.session, filter, opts...)
}

// DeleteOneByID 在事务中根据ID删除单个文档
func (tr *TransactionRepository) DeleteOneByID(id interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return tr.Repository.DeleteOneByID(tr.session, id, opts...)
}

// DeleteMany 在事务中删除多个文档
func (tr *TransactionRepository) DeleteMany(filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return tr.Repository.DeleteMany(tr.session, filter, opts...)
}

// CountDocuments 在事务中统计文档数量
func (tr *TransactionRepository) CountDocuments(filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return tr.Repository.CountDocuments(tr.session, filter, opts...)
}

// Aggregate 在事务中执行聚合查询
func (tr *TransactionRepository) Aggregate(pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return tr.Repository.Aggregate(tr.session, pipeline, opts...)
}

// AggregateAll 在事务中执行聚合查询并解码结果
func (tr *TransactionRepository) AggregateAll(pipeline interface{}, results interface{}, opts ...*options.AggregateOptions) error {
	return tr.Repository.AggregateAll(tr.session, pipeline, results, opts...)
}

// BulkWrite 在事务中执行批量写操作
func (tr *TransactionRepository) BulkWrite(models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return tr.Repository.BulkWrite(tr.session, models, opts...)
}

// Session 获取会话上下文
func (tr *TransactionRepository) Session() mongo.SessionContext {
	return tr.session
}

// TransactionHelper 事务辅助函数
type TransactionHelper struct {
	tm *TransactionManager
}

// NewTransactionHelper 创建事务辅助器
func NewTransactionHelper(client *Client) *TransactionHelper {
	return &TransactionHelper{
		tm: NewTransactionManager(client),
	}
}

// ExecuteInTransaction 在事务中执行多个操作
func (th *TransactionHelper) ExecuteInTransaction(ctx context.Context, operations []func(*TransactionRepository) error, opts ...*TransactionOptions) error {
	return th.tm.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		for i, operation := range operations {
			// 为每个操作创建一个临时仓储（这里需要根据实际需求调整）
			// 由于我们不知道具体的集合，这里提供一个通用的方法
			if err := operation(nil); err != nil {
				th.tm.logger.Error("Transaction operation failed", "operation", i, "error", err)
				return nil, fmt.Errorf("operation %d failed: %w", i, err)
			}
		}
		return nil, nil
	}, opts...)
}

// ExecuteWithRetry 带重试的事务执行
func (th *TransactionHelper) ExecuteWithRetry(ctx context.Context, fn TransactionFunc, maxRetries int, opts ...*TransactionOptions) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		err := th.tm.WithTransaction(ctx, fn, opts...)
		if err == nil {
			return nil
		}

		lastErr = err
		th.tm.logger.Warn("Transaction attempt failed", "attempt", i+1, "error", err)

		// 如果不是最后一次重试，等待一段时间
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}

	th.tm.logger.Error("All transaction attempts failed", "attempts", maxRetries+1, "lastError", lastErr)
	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries+1, lastErr)
}