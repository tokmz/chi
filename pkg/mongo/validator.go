package mongo

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Validator 文档验证器接口
type Validator interface {
	Validate(document interface{}) error
	ValidateUpdate(update interface{}) error
	GetSchema() any
}

// SchemaValidator 基于Schema的验证器
type SchemaValidator struct {
	schema map[string]interface{}
	logger Logger
}

// NewSchemaValidator 创建Schema验证器
func NewSchemaValidator(schema map[string]interface{}, logger Logger) *SchemaValidator {
	return &SchemaValidator{
		schema: schema,
		logger: logger,
	}
}

// Validate 验证文档
func (sv *SchemaValidator) Validate(document interface{}) error {
	if sv.schema == nil {
		return nil // 没有schema则跳过验证
	}

	// 将文档转换为map
	docMap, err := sv.toMap(document)
	if err != nil {
		return fmt.Errorf("failed to convert document to map: %w", err)
	}

	// 验证必填字段
	if err := sv.validateRequired(docMap); err != nil {
		return err
	}

	// 验证字段类型
	if err := sv.validateTypes(docMap); err != nil {
		return err
	}

	// 验证字段值
	if err := sv.validateValues(docMap); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate 验证更新文档
func (sv *SchemaValidator) ValidateUpdate(update interface{}) error {
	if sv.schema == nil {
		return nil
	}

	updateMap, err := sv.toMap(update)
	if err != nil {
		return fmt.Errorf("failed to convert update to map: %w", err)
	}

	// 验证更新操作
	return sv.validateUpdateOperations(updateMap)
}

// GetSchema 获取Schema
func (sv *SchemaValidator) GetSchema() interface{} {
	return sv.schema
}

// toMap 将interface{}转换为map[string]interface{}
func (sv *SchemaValidator) toMap(data interface{}) (map[string]interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return v, nil
	case bson.M:
		return map[string]interface{}(v), nil
	case bson.D:
		return v.Map(), nil
	default:
		// 使用反射转换结构体
		return sv.structToMap(data)
	}
}

// structToMap 将结构体转换为map
func (sv *SchemaValidator) structToMap(data interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %T", data)
	}

	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过未导出的字段
		if !fieldValue.CanInterface() {
			continue
		}

		// 获取bson标签
		bsonTag := field.Tag.Get("bson")
		if bsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if bsonTag != "" {
			parts := strings.Split(bsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		result[fieldName] = fieldValue.Interface()
	}

	return result, nil
}

// validateRequired 验证必填字段
func (sv *SchemaValidator) validateRequired(docMap map[string]interface{}) error {
	required, ok := sv.schema["required"]
	if !ok {
		return nil
	}

	requiredFields, ok := required.([]interface{})
	if !ok {
		return nil
	}

	for _, field := range requiredFields {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}

		if _, exists := docMap[fieldName]; !exists {
			return fmt.Errorf("required field '%s' is missing", fieldName)
		}
	}

	return nil
}

// validateTypes 验证字段类型
func (sv *SchemaValidator) validateTypes(docMap map[string]interface{}) error {
	properties, ok := sv.schema["properties"]
	if !ok {
		return nil
	}

	propsMap, ok := properties.(map[string]interface{})
	if !ok {
		return nil
	}

	for fieldName, value := range docMap {
		propSchema, exists := propsMap[fieldName]
		if !exists {
			continue
		}

		propMap, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		expectedType, ok := propMap["type"]
		if !ok {
			continue
		}

		if err := sv.validateFieldType(fieldName, value, expectedType); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldType 验证字段类型
func (sv *SchemaValidator) validateFieldType(fieldName string, value interface{}, expectedType interface{}) error {
	typeStr, ok := expectedType.(string)
	if !ok {
		return nil
	}

	actualType := sv.getValueType(value)
	if !sv.isTypeCompatible(actualType, typeStr) {
		return fmt.Errorf("field '%s' expected type '%s', got '%s'", fieldName, typeStr, actualType)
	}

	return nil
}

// getValueType 获取值的类型
func (sv *SchemaValidator) getValueType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "integer"
	case float32, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}, bson.M:
		return "object"
	case primitive.ObjectID:
		return "objectId"
	case time.Time:
		return "date"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// isTypeCompatible 检查类型兼容性
func (sv *SchemaValidator) isTypeCompatible(actualType, expectedType string) bool {
	if actualType == expectedType {
		return true
	}

	// 特殊兼容性规则
	switch expectedType {
	case "number":
		return actualType == "integer" || actualType == "number"
	case "integer":
		return actualType == "integer"
	default:
		return false
	}
}

// validateValues 验证字段值
func (sv *SchemaValidator) validateValues(docMap map[string]interface{}) error {
	properties, ok := sv.schema["properties"]
	if !ok {
		return nil
	}

	propsMap, ok := properties.(map[string]interface{})
	if !ok {
		return nil
	}

	for fieldName, value := range docMap {
		propSchema, exists := propsMap[fieldName]
		if !exists {
			continue
		}

		propMap, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		if err := sv.validateFieldValue(fieldName, value, propMap); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldValue 验证字段值
func (sv *SchemaValidator) validateFieldValue(fieldName string, value interface{}, propSchema map[string]interface{}) error {
	// 验证最小值
	if minimum, ok := propSchema["minimum"]; ok {
		if err := sv.validateMinimum(fieldName, value, minimum); err != nil {
			return err
		}
	}

	// 验证最大值
	if maximum, ok := propSchema["maximum"]; ok {
		if err := sv.validateMaximum(fieldName, value, maximum); err != nil {
			return err
		}
	}

	// 验证最小长度
	if minLength, ok := propSchema["minLength"]; ok {
		if err := sv.validateMinLength(fieldName, value, minLength); err != nil {
			return err
		}
	}

	// 验证最大长度
	if maxLength, ok := propSchema["maxLength"]; ok {
		if err := sv.validateMaxLength(fieldName, value, maxLength); err != nil {
			return err
		}
	}

	// 验证枚举值
	if enum, ok := propSchema["enum"]; ok {
		if err := sv.validateEnum(fieldName, value, enum); err != nil {
			return err
		}
	}

	return nil
}

// validateMinimum 验证最小值
func (sv *SchemaValidator) validateMinimum(fieldName string, value interface{}, minimum interface{}) error {
	valueNum, ok1 := sv.toNumber(value)
	minNum, ok2 := sv.toNumber(minimum)
	if !ok1 || !ok2 {
		return nil
	}

	if valueNum < minNum {
		return fmt.Errorf("field '%s' value %v is less than minimum %v", fieldName, value, minimum)
	}

	return nil
}

// validateMaximum 验证最大值
func (sv *SchemaValidator) validateMaximum(fieldName string, value interface{}, maximum interface{}) error {
	valueNum, ok1 := sv.toNumber(value)
	maxNum, ok2 := sv.toNumber(maximum)
	if !ok1 || !ok2 {
		return nil
	}

	if valueNum > maxNum {
		return fmt.Errorf("field '%s' value %v is greater than maximum %v", fieldName, value, maximum)
	}

	return nil
}

// validateMinLength 验证最小长度
func (sv *SchemaValidator) validateMinLength(fieldName string, value interface{}, minLength interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	minLen, ok := sv.toInt(minLength)
	if !ok {
		return nil
	}

	if len(str) < minLen {
		return fmt.Errorf("field '%s' length %d is less than minimum %d", fieldName, len(str), minLen)
	}

	return nil
}

// validateMaxLength 验证最大长度
func (sv *SchemaValidator) validateMaxLength(fieldName string, value interface{}, maxLength interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	maxLen, ok := sv.toInt(maxLength)
	if !ok {
		return nil
	}

	if len(str) > maxLen {
		return fmt.Errorf("field '%s' length %d is greater than maximum %d", fieldName, len(str), maxLen)
	}

	return nil
}

// validateEnum 验证枚举值
func (sv *SchemaValidator) validateEnum(fieldName string, value interface{}, enum interface{}) error {
	enumSlice, ok := enum.([]interface{})
	if !ok {
		return nil
	}

	for _, enumValue := range enumSlice {
		if reflect.DeepEqual(value, enumValue) {
			return nil
		}
	}

	return fmt.Errorf("field '%s' value %v is not in enum %v", fieldName, value, enum)
}

// validateUpdateOperations 验证更新操作
func (sv *SchemaValidator) validateUpdateOperations(updateMap map[string]interface{}) error {
	for operator, operand := range updateMap {
		switch operator {
		case "$set", "$unset", "$inc", "$mul", "$rename", "$min", "$max", "$currentDate":
			if err := sv.validateUpdateOperator(operator, operand); err != nil {
				return err
			}
		default:
			// 对于不认识的操作符，记录警告但不阻止
			sv.logger.Warn("Unknown update operator", "operator", operator)
		}
	}

	return nil
}

// validateUpdateOperator 验证更新操作符
func (sv *SchemaValidator) validateUpdateOperator(operator string, operand interface{}) error {
	operandMap, ok := operand.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid operand for %s: expected object", operator)
	}

	switch operator {
	case "$set":
		return sv.validateTypes(operandMap)
	case "$inc", "$mul":
		return sv.validateNumericOperations(operandMap)
	default:
		return nil
	}
}

// validateNumericOperations 验证数值操作
func (sv *SchemaValidator) validateNumericOperations(operandMap map[string]interface{}) error {
	for fieldName, value := range operandMap {
		if !sv.isNumeric(value) {
			return fmt.Errorf("field '%s' in numeric operation must be a number, got %T", fieldName, value)
		}
	}
	return nil
}

// isNumeric 检查值是否为数值类型
func (sv *SchemaValidator) isNumeric(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}

// toNumber 转换为数值
func (sv *SchemaValidator) toNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// toInt 转换为整数
func (sv *SchemaValidator) toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// ValidatedRepository 带验证的仓储
type ValidatedRepository struct {
	*Repository
	validator Validator
}

// NewValidatedRepository 创建带验证的仓储
func NewValidatedRepository(repo *Repository, validator Validator) *ValidatedRepository {
	return &ValidatedRepository{
		Repository: repo,
		validator:  validator,
	}
}

// InsertOne 插入单个文档（带验证）
func (vr *ValidatedRepository) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	if err := vr.validator.Validate(document); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return vr.Repository.InsertOne(ctx, document, opts...)
}

// InsertMany 插入多个文档（带验证）
func (vr *ValidatedRepository) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	for i, doc := range documents {
		if err := vr.validator.Validate(doc); err != nil {
			return nil, fmt.Errorf("validation failed for document %d: %w", i, err)
		}
	}
	return vr.Repository.InsertMany(ctx, documents, opts...)
}

// UpdateOne 更新单个文档（带验证）
func (vr *ValidatedRepository) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if err := vr.validator.ValidateUpdate(update); err != nil {
		return nil, fmt.Errorf("update validation failed: %w", err)
	}
	return vr.Repository.UpdateOne(ctx, filter, update, opts...)
}

// UpdateMany 更新多个文档（带验证）
func (vr *ValidatedRepository) UpdateMany(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if err := vr.validator.ValidateUpdate(update); err != nil {
		return nil, fmt.Errorf("update validation failed: %w", err)
	}
	return vr.Repository.UpdateMany(ctx, filter, update, opts...)
}

// ReplaceOne 替换单个文档（带验证）
func (vr *ValidatedRepository) ReplaceOne(ctx context.Context, filter, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	if err := vr.validator.Validate(replacement); err != nil {
		return nil, fmt.Errorf("replacement validation failed: %w", err)
	}
	return vr.Repository.ReplaceOne(ctx, filter, replacement, opts...)
}

// GetValidator 获取验证器
func (vr *ValidatedRepository) GetValidator() Validator {
	return vr.validator
}

// SetValidator 设置验证器
func (vr *ValidatedRepository) SetValidator(validator Validator) {
	vr.validator = validator
}
