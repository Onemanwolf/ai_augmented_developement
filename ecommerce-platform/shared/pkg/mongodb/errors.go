package mongodb

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// Common MongoDB errors.
var (
	ErrNotFound      = errors.New("document not found")
	ErrDuplicateKey  = errors.New("duplicate key error")
	ErrInvalidID     = errors.New("invalid document ID")
	ErrNoTransaction = errors.New("operation requires a transaction")
)

// IsNotFoundError checks if the error indicates a document was not found.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, mongo.ErrNoDocuments) || errors.Is(err, ErrNotFound)
}

// IsDuplicateKeyError checks if the error is a duplicate key error.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	// Check for our wrapped error
	if errors.Is(err, ErrDuplicateKey) {
		return true
	}

	// Check for MongoDB write exception with duplicate key code
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, we := range writeErr.WriteErrors {
			// 11000 is the MongoDB error code for duplicate key
			if we.Code == 11000 {
				return true
			}
		}
	}

	return false
}

// IsTransientError checks if the error is transient and the operation can be retried.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Check for command errors that are transient
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		// Check for transient transaction error labels
		if cmdErr.HasErrorLabel("TransientTransactionError") {
			return true
		}
	}

	return false
}

// WrapNotFound wraps mongo.ErrNoDocuments with our ErrNotFound.
func WrapNotFound(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	return err
}
