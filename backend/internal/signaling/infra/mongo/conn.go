package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	mongodrv "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	once    sync.Once
	db      *mongodrv.Database
	connErr error
)

func Init(dsn, dbName string, pool uint64) {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		// TODO: Log error here
		client, err := mongodrv.Connect(ctx,
			options.Client().
				ApplyURI(dsn).
				SetAppName("vidcall-test").
				SetMaxPoolSize(pool),
		)

		if err != nil {
			fmt.Println("MongoDB Connection Error \n\n %w", err)
		}

		db = client.Database(dbName)

	})
}

func DB() *mongodrv.Database { return db }
