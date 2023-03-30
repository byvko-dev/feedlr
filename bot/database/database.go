package database

import (
	"context"
	"errors"
	"time"

	p "github.com/byvko-dev/feedlr/prisma/client"
)

type Database struct {
	client *p.PrismaClient
}

var Client = &Database{
	client: p.NewClient(),
}

func (d *Database) Connect() error {
	return d.client.Connect()
}

func (d *Database) Close() {
	d.client.Disconnect()
}

func (d *Database) CreateGuild(id string, name string) (*p.GuildModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	guild, err := d.client.Guild.CreateOne(
		p.Guild.ID.Set(id),
		p.Guild.Name.Set(name),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return guild, nil
}

func (d *Database) GetGuild(id string) (*p.GuildModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	guild, err := d.client.Guild.FindUnique(p.Guild.ID.Equals(id)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return guild, nil
}

func (d *Database) GetOrCreateGuild(id string, name string) (*p.GuildModel, error) {
	guild, err := d.GetGuild(id)
	if err != nil {
		if errors.Is(err, p.ErrNotFound) {
			return d.CreateGuild(id, name)
		}
		return nil, err
	}

	// Update guild name if it has changed
	defer func() {
		if guild.Name != name {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			_, err := d.client.Guild.FindUnique(p.Guild.ID.Equals(id)).Update(p.Guild.Name.Set(name)).Exec(ctx)
			if err != nil {
				return
			}
		}
	}()

	return guild, nil
}

func (d *Database) CreateFeed(url string) (*p.FeedModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	feed, err := d.client.Feed.CreateOne(
		p.Feed.URL.Set(url),
		p.Feed.LastFetch.Set(time.Now()),
	).With(p.Feed.Webhooks.Fetch()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

func (d *Database) GetFeed(id string) (*p.FeedModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	feed, err := d.client.Feed.FindUnique(p.Feed.ID.Equals(id)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

func (d *Database) GetManyFeeds(ids []string) ([]p.FeedModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	feeds, err := d.client.Feed.FindMany(p.Feed.ID.In(ids)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return feeds, nil
}

func (d *Database) FindFeedByURL(url string) (*p.FeedModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	feed, err := d.client.Feed.FindFirst(
		p.Feed.URL.Equals(url),
	).With(p.Feed.Webhooks.Fetch()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

func (d *Database) GetOrCreateFeed(url string) (*p.FeedModel, error) {
	feed, err := d.FindFeedByURL(url)
	if err != nil {
		if errors.Is(err, p.ErrNotFound) {
			return d.CreateFeed(url)
		}
		return nil, err
	}
	return feed, nil
}

func (d *Database) CreateWebhook(feedID, guildD, channelID, name, id, token string) (*p.WebhookModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	webhook, err := d.client.Webhook.CreateOne(
		p.Webhook.Name.Set(name),
		p.Webhook.Token.Set(token),
		p.Webhook.ExternalID.Set(id),
		p.Webhook.ChannelID.Set(channelID),
		p.Webhook.Guild.Link(p.Guild.ID.Equals(guildD)),
		p.Webhook.Feed.Link(p.Feed.ID.Equals(feedID)),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

func (d *Database) GetWebhook(id string) (*p.WebhookModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	webhook, err := d.client.Webhook.FindUnique(p.Webhook.ID.Equals(id)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

func (d *Database) FindWebhooksByChannelID(id string) ([]p.WebhookModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	webhooks, err := d.client.Webhook.FindMany(
		p.Webhook.ChannelID.Equals(id),
	).With(p.Webhook.Feed.Fetch()).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (d *Database) DeleteWebhooksByChannelID(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := d.client.Webhook.FindMany(
		p.Webhook.ChannelID.Equals(id),
	).Delete().Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) DeleteWebhook(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := d.client.Webhook.FindUnique(p.Webhook.ID.Equals(id)).Delete().Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
