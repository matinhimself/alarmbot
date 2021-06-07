package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// Check if error is for duplicated unique id in collection
func IsMongoDupError(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}
