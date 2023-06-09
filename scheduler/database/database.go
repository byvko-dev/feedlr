package database

import (
	"context"
	"time"

	p "github.com/byvko-dev/feedlr/prisma/client"
)

type Database struct {
	client *p.PrismaClient
}

var db = &Database{
	client: p.NewClient(),
}

func GetDatabase() *Database {
	return db
}

func (d *Database) Connect() error {
	return d.client.Connect()
}

func (d *Database) Close() {
	d.client.Disconnect()
}

func (d *Database) GetAllFeeds() ([]p.FeedModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	feed, err := d.client.Feed.FindMany().With(p.Feed.Webhooks.Fetch()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

func (d *Database) GetFeedByID(id string) (*p.FeedModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	feed, err := d.client.Feed.FindUnique(p.Feed.ID.Equals(id)).With(p.Feed.Webhooks.Fetch()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

func (d *Database) UpdateFeedsLastFetched(timestamp time.Time, ids ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := d.client.Feed.FindMany(p.Feed.ID.In(ids)).Update(p.Feed.LastFetch.Set(timestamp)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
