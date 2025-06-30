package mongorepo

import (
	"context"
	"time"
	"vidcall/internal/signaling/domain"
	"vidcall/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type roomDoc struct {
	RoomID   string    `bson:"roomID"`
	HostID   string    `bson:"hostID"`
	Pin      string    `bson:"pin"`
	Date     time.Time `bson:"date"`
	Duration string    `bson:"duration"`
}

func toRoomDoc(r domain.Room) roomDoc {
	return roomDoc{
		RoomID:   r.RoomID,
		HostID:   r.HostID,
		Pin:      r.Pin,
		Date:     r.Date,
		Duration: r.Duration.String(),
	}
}

func fromRoomDoc(rd roomDoc) domain.Room {
	dur, _ := time.ParseDuration(rd.Duration)

	return domain.Room{
		RoomID:   rd.RoomID,
		HostID:   rd.HostID,
		Pin:      rd.Pin,
		Date:     rd.Date,
		Duration: dur,
	}
}

func CreateRoomDoc(ctx context.Context, db *mongo.Database, room domain.Room) {
	log := logger.GetLog(ctx).With("layer", "repo", "roomID", room.RoomID)

	col := db.Collection("rooms")

	// TODO: do I need this???
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := col.InsertOne(opCtx, toRoomDoc(room))
	if err != nil {
		log.Warn("Unable to insert document")
		return
	}
}

func GetRoomDoc(ctx context.Context, db *mongo.Database, roomID string) *domain.Room {
	log := logger.GetLog(ctx).With("layer", "repo", "roomID", roomID)

	col := db.Collection("rooms")

	opCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var d roomDoc
	err := col.FindOne(opCtx, bson.M{"RoomID": roomID}).Decode(d)

	if err != nil {
		log.Warn("Unable to find document")
		return nil
	}

	room := fromRoomDoc(d)

	return &room
}

func RemoveRoomDoc(ctx context.Context, db *mongo.Database, roomID string) {}
