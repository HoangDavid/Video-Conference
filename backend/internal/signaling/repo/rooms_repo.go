package repo

import (
	"context"
	"fmt"
	"time"
	"vidcall/internal/signaling/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type roomDoc struct {
	ID       primitive.ObjectID
	RoomID   string
	HostID   string
	Members  map[string]domain.Member
	Pin      string // already hashed
	Date     time.Time
	Duration time.Duration
}

// TODO: memberDoc maybe

func toDoc(r domain.Room) roomDoc {
	return roomDoc{
		RoomID:   r.RoomID,
		HostID:   r.HostID,
		Members:  r.Members,
		Pin:      r.Pin,
		Date:     r.Date,
		Duration: r.Duration,
	}
}

func fromDoc(rd roomDoc) domain.Room {
	return domain.Room{
		RoomID:   rd.RoomID,
		HostID:   rd.HostID,
		Members:  rd.Members,
		Pin:      rd.Pin,
		Date:     rd.Date,
		Duration: rd.Duration,
	}
}

func CreateRoomDoc(ctx context.Context, db *mongo.Database, room domain.Room) {
	col := db.Collection("rooms")

	// TODO: add the timeout time as constant
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// TODO: add error logging here
	_, err := col.InsertOne(opCtx, toDoc(room))

	if err != nil {
		fmt.Println("MongoDB Room Creation Error \n\n %w", err)
	}
}

func GetRoomDoc(ctx context.Context, db *mongo.Database, roomID string) *domain.Room {
	col := db.Collection("rooms")

	// TODO: add timeout as context
	opCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var d roomDoc

	err := col.FindOne(opCtx, bson.M{"RoomID": roomID}).Decode(d)

	// TODO: Add error logging here
	if err != nil {
		fmt.Println("MongoDB Room Find Error \n\n %w", err)
	}

	room := fromDoc(d)

	return &room
}

func RemoveRoomDoc(ctx context.Context, db *mongo.Database, roomID string) {}

func AddMember(ctx context.Context, db *mongo.Database, roomID string) {}

func RemoveMemeber(ctx context.Context, db *mongo.Database, peerID string) {}
