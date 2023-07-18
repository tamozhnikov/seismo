package mongodb

import (
	"context"
	"fmt"
	"path"
	"seismo/provider"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	msgCollName = "messages"
)

// Adapter provides interaction with a MONGODB database.
type Adapter struct {
	//connStr specifies a connection string.
	connStr string

	//dbName specifies the name of the database.
	dbName string

	client mongo.Client
}

// Connect opens a new connection to a database using "connStr" as the connection string
// and initializes the adapter.
func (a *Adapter) Connect(ctx context.Context, connStr string) error {
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		return fmt.Errorf("Connect: error: %w", err)
	}
	a.connStr = connStr
	a.dbName = path.Base(connStr)
	a.client = *c
	return nil
}

// Close closes the opened connection.
func (a *Adapter) Close(ctx context.Context) error {
	err := a.client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("Close: error: %w", err)
	}
	return nil
}

// SaveMsg saves messages in the connected database.
func (a *Adapter) SaveMsg(ctx context.Context, msgs []provider.Message) error {
	coll := a.client.Database(a.dbName).Collection(msgCollName)
	mi := make([]interface{}, 0, len(msgs))
	for _, m := range msgs {
		mi = append(mi, m)
	}
	_, err := coll.InsertMany(ctx, mi)
	if err != nil {
		return fmt.Errorf("SaveMsg: error: %w", err)
	}

	return nil
}

// GetLastTime returns the focus time of the last saved message for specified "sourceId".
func (a *Adapter) GetLastTime(ctx context.Context, sourceId string) (time.Time, error) {
	coll := a.client.Database(a.dbName).Collection(msgCollName)

	//db.messages.aggregate([{$match: {sourceid: "pseudo_1"}},  {$group : { _id: "$source_id", max_magn: {$max : "$focustime"}}}])
	cur, err := coll.Aggregate(ctx, mongo.Pipeline{
		{{"$match", bson.D{{"source_id", sourceId}}}},
		{{"$group", bson.D{{"_id", "$source_id"}, {"time", bson.D{{"$max", "$focus_time"}}}}}},
	})

	if err != nil {
		return time.Time{}, fmt.Errorf("GetLastTime: error: %w", err)
	}
	var res []bson.M
	if err = cur.All(ctx, &res); err != nil {
		return time.Time{}, fmt.Errorf("GetLastTime: error: %w", err)
	}

	if len(res) < 1 {
		return time.Time{}, nil
	}

	i := res[0]["time"]
	t, ok := i.(primitive.DateTime)
	if !ok {
		return time.Time{}, fmt.Errorf("GetLastTime: unexpected type for time %v; DateTime type is exected", res[0]["time"])
	}

	return t.Time(), nil
}
