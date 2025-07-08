package redisrepo

import (
	"context"
	"vidcall/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// Redis Set to keep room memberships:
// Redis Hash to keep room state (ex: (un)muted states)

// Helper funtions
func roomKey(roomID string) string   { return "rooms:" + roomID }
func memberKey(roomID string) string { return "rooms:" + roomID + ":members" }
func metaKey(roomID string) string   { return "rooms:" + roomID + ":meta" }

// Create room
func CreateRoom(ctx context.Context, rdb *redis.Client, roomID string) error {

	log := logger.GetLog(ctx).With("layer", "repo", "service", "redis", "roomID", roomID)

	pipe := rdb.TxPipeline()
	if err := pipe.SAdd(ctx, memberKey(roomID), "__init__").Err(); err != nil {
		log.Error("unable to create room key")
		return err
	}

	if err := pipe.HSet(ctx, metaKey(roomID), "__init__", "{}").Err(); err != nil {
		log.Error("unable to add room metadata")
		return err
	}

	return nil
}

// Delete room
func DelRoom(ctx context.Context, rdb *redis.Client, roomID string) error {

	log := logger.GetLog(ctx).With("layer", "repo", "service", "redis", "roomID", roomID)

	if err := rdb.Del(ctx, memberKey(roomID), metaKey(roomID)).Err(); err != nil {
		log.Error("unable to remove room")
		return err
	}
	return nil
}

// Add Member
func AddMember(ctx context.Context, rdb *redis.Client, roomID string, peerID string) error {

	log := logger.GetLog(ctx).With("layer", "repo", "service", "redis", "roomID", roomID)

	pipe := rdb.TxPipeline()

	k := memberKey(roomID)
	if err := pipe.SRem(ctx, k, "__init__").Err(); err != nil {
		log.Error("unable to remove member")
		return err
	}

	if err := pipe.SAdd(ctx, k, peerID).Err(); err != nil {
		log.Error("unable to add member")
		return err
	}

	return nil

}

// Delete Memeber
func DelMember(ctx context.Context, rdb *redis.Client, roomID string, peerID string) (empty bool, err error) {

	log := logger.GetLog(ctx).With("layer", "repo", "service", "redis", "roomID", roomID)

	if err := rdb.SRem(ctx, memberKey(roomID), peerID).Err(); err != nil {
		log.Error("unable to remove member")
		return false, err
	}
	room_count, err := rdb.SCard(ctx, memberKey(roomID)).Result()

	return (room_count == 0 && err != nil), err
}

// Room Subcription
func Subcribe(ctx context.Context, rdb *redis.Client, roomID string) *redis.PubSub {
	sub := rdb.Subscribe(ctx, roomKey(roomID))
	return sub
}

// Broadcasting to members
func Broadcast(ctx context.Context, rdb *redis.Client, roomID string, payload any) error {

	log := logger.GetLog(ctx).With("layer", "repo", "service", "redis", "roomID", roomID)

	if err := rdb.Publish(ctx, roomKey(roomID), payload).Err(); err != nil {
		log.Error("unable to publish")
	}

	return nil
}

// TODO: do room meta for audio/video/translation/subtitle tracks
