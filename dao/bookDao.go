package dao

import (
	"BookManagement/model"
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookDao struct {
	Server     string
	Database   string
	Collection string
}

var Collection *mongo.Collection
var ctx = context.TODO()

func (b *BookDao) Connect() {
	clientOptions := options.Client().ApplyURI(b.Server)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database(b.Database).Collection(b.Collection)
	fmt.Println("connected")
}

func (b *BookDao) Insert(bm model.Book) error {
	_, err := Collection.InsertOne(ctx, bm)

	if err != nil {
		fmt.Println("not inserted")
	}
	return nil
}

func (b *BookDao) FindByBookName(bookId string) ([]*model.Book, error) {
	var books []*model.Book

	cur, err := Collection.Find(ctx, bson.D{primitive.E{Key: "name", Value: "the cook book"}})

	if err != nil {
		return books, errors.New("unable to query db")
	}

	for cur.Next(ctx) {
		var b model.Book

		err := cur.Decode(&b)

		if err != nil {
			return books, err
		}

		books = append(books, &b)
	}

	if err := cur.Err(); err != nil {
		return books, err
	}

	cur.Close(ctx)

	if len(books) == 0 {
		return books, mongo.ErrNoDocuments
	}

	return books, nil
}

func (b *BookDao) DeleteBook(name string) error {
	filter := bson.D{primitive.E{Key: "name", Value: name}}

	res, err := Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("no books were deleted")
	}

	return nil
}

func (b *BookDao) UpdateBook(name string, book model.Book) error {
	filter := bson.D{primitive.E{Key: "name", Value: name}}

	update := bson.D{primitive.E{Key: "$set", Value: book}}

	c := &model.Book{}
	return Collection.FindOneAndUpdate(ctx, filter, update).Decode(c)
}
