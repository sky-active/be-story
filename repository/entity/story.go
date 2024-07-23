package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type Story struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    string             `bson:"userID"`
	Content   string             `bson:"content"`
	Views     int                `bson:"views"`
	ExpiresAt int64              `bson:"expiresAt"` // UNIX timestamp
}
