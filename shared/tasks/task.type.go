package tasks

// Task is a task to be executed by the worker
type Task struct {
	FeedID      string `json:"feed_id"`
	WebhookURL  string `json:"webhook_url"`
	WebhookName string `json:"webhook_name"`
	Post        Post   `json:"post"`
}

// RSS feed post
type Post struct {
	Link        string `json:"link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	PubDate     string `json:"pub_date"`
}
