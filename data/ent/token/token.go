// Code generated by ent, DO NOT EDIT.

package token

import (
	"time"
)

const (
	// Label holds the string label denoting the token type in the database.
	Label = "token"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldToken holds the string denoting the token field in the database.
	FieldToken = "token"
	// FieldSource holds the string denoting the source field in the database.
	FieldSource = "source"
	// FieldExpiredAt holds the string denoting the expired_at field in the database.
	FieldExpiredAt = "expired_at"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// Table holds the table name of the token in the database.
	Table = "sys_tokens"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "sys_tokens"
	// OwnerInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	OwnerInverseTable = "sys_users"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "user_token"
)

// Columns holds all SQL columns for token fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldUserID,
	FieldToken,
	FieldSource,
	FieldExpiredAt,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "sys_tokens"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"user_token",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
)
