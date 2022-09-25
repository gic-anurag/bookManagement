package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Book struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name  string             `bson:"name,omitempty" json:"name,omitempty"`
	Cost  int                `bson:"cost,omitempty" json:"cost,omitempty`
	Pages int                `bson:"pages,omitempty" json:"pages,omitempty`
}
