package mongo

import "errors"

// 定义MongoDB模块的错误
var (
	// 配置相关错误
	ErrInvalidURI      = errors.New("invalid MongoDB URI")
	ErrInvalidDatabase = errors.New("invalid database name")
	ErrInvalidPoolSize = errors.New("max pool size must be greater than min pool size")

	// 连接相关错误
	ErrConnectionFailed = errors.New("failed to connect to MongoDB")
	ErrConnectionClosed = errors.New("MongoDB connection is closed")
	ErrPingFailed       = errors.New("failed to ping MongoDB")

	// 操作相关错误
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidObjectID  = errors.New("invalid ObjectID")
	ErrInvalidFilter    = errors.New("invalid filter")
	ErrInvalidUpdate    = errors.New("invalid update")
	ErrInvalidDocument  = errors.New("invalid document")

	// 事务相关错误
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrTransactionAborted  = errors.New("transaction aborted")
	ErrTransactionCommited = errors.New("transaction already committed")

	// 验证相关错误
	ErrValidationFailed = errors.New("document validation failed")
	ErrSchemaNotFound   = errors.New("schema not found")

	// 索引相关错误
	ErrIndexCreationFailed = errors.New("failed to create index")
	ErrIndexNotFound       = errors.New("index not found")

	// 集合相关错误
ErrCollectionNotFound = errors.New("collection not found")
ErrCollectionExists   = errors.New("collection already exists")

// 日志相关错误
ErrInvalidLogLevel = errors.New("invalid log level")
ErrNoLogOutput     = errors.New("no log output configured")
ErrInvalidLogFile  = errors.New("invalid log file configuration")
)