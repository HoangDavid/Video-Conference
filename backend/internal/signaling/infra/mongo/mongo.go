package mongo

import (
	"context"
	"sync"
	"time"
	"vidcall/pkg/logger"

	mongodrv "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	once sync.Once
	db   *mongodrv.Database
)

func Init(dsn, dbName string, pool uint64) {

	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		client, err := mongodrv.Connect(ctx,
			options.Client().
				ApplyURI(dsn).
				SetAppName("vidcall-test").
				SetMaxPoolSize(pool),
		)

		log := logger.GetLog(ctx).With("layer", "infra", "service", "mongodb")
		if err != nil {
			log.Error("Unable to connect to MongoDB")
			return
		}

		db = client.Database(dbName)

	})
}

func DB() *mongodrv.Database { return db }
