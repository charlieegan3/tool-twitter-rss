package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// NewEntry is a job that gets a user's twitter feed for a day and creates an rss entry for it
type NewEntry struct {
	ScheduleOverride string
	Endpoint         string

	AccessToken       string
	AccessTokenSecret string
	ConsumerKey       string
	ConsumerSecret    string
}

func (n *NewEntry) Name() string {
	return "new-entry"
}

func (n *NewEntry) Run(ctx context.Context) error {
	doneCh := make(chan bool)
	errCh := make(chan error)

	go func() {
		config := oauth1.NewConfig(n.ConsumerKey, n.ConsumerSecret)
		token := oauth1.NewToken(n.AccessToken, n.AccessTokenSecret)
		httpClient := config.Client(oauth1.NoContext, token)
		twitterClient := twitter.NewClient(httpClient)

		selectedTweets := []twitter.Tweet{}
		selectedDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

		var maxID int64

		for {
			excludeReplies := true

			var tweets []twitter.Tweet
			var err error

			if maxID == 0 {
				tweets, _, err = twitterClient.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
					Count:          50,
					ExcludeReplies: &excludeReplies,
				})
			} else {
				tweets, _, err = twitterClient.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
					Count:          50,
					ExcludeReplies: &excludeReplies,
					MaxID:          maxID,
				})
			}
			if err != nil {
				errCh <- fmt.Errorf("failed to get tweets: %s", err)
				return
			}

			stop := true
			for i, tweet := range tweets {
				createdAt, err := tweet.CreatedAtTime()
				if err != nil {
					continue
				}
				if createdAt.Format("2006-01-02") == selectedDate {
					selectedTweets = append([]twitter.Tweet{tweet}, selectedTweets...)
				}

				if i == len(tweets)-1 {
					maxID = tweet.ID - 1
					if createdAt.Format("2006-01-02") == selectedDate {
						stop = false
					}
				}
			}
			if stop {
				break
			}
		}

		log.Println("found", len(selectedTweets), "tweets for", selectedDate)

		const template = `
<p style="font-size: 0.8rem; font-family: sans-serif;">
<img style="border-radius: 100rem; margin-bottom: -0.5rem; width: 1.7rem;"
     src="%s">
%s
(@%s)
<a href="https://twitter.com/%s/status/%s">
%s
</a>
<p>

<blockquote style="font-family: sans-serif;">
<p>%s</p>
</blockquote>
<hr/>
`

		html := ""
		for _, t := range selectedTweets {
			createdAt, err := t.CreatedAtTime()
			if err != nil {
				continue
			}
			html += fmt.Sprintf(
				template,
				t.User.ProfileImageURLHttps,
				t.User.Name,
				t.User.ScreenName,
				t.User.ScreenName,
				t.IDStr,
				createdAt.Format("15:04"),
				t.Text,
			)
		}

		datab := []map[string]string{
			{
				"title": fmt.Sprintf("Tweets on %s", selectedDate),
				"body":  html,
				"url":   "",
			},
		}

		b, err := json.Marshal(datab)
		if err != nil {
			errCh <- fmt.Errorf("failed to form new item JSON: %s", err)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", n.Endpoint, bytes.NewBuffer(b))
		if err != nil {
			errCh <- fmt.Errorf("failed to build request for new item: %s", err)
			return
		}

		req.Header.Add("Content-Type", "application/json; charset=utf-8")

		resp, err := client.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("failed to send request for new item: %s", err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			errCh <- fmt.Errorf("failed to send request: non 200OK response")
			return
		}

		doneCh <- true
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-errCh:
		return fmt.Errorf("job failed with error: %s", e)
	case <-doneCh:
		return nil
	}
}

func (n *NewEntry) Timeout() time.Duration {
	return 30 * time.Second
}

func (n *NewEntry) Schedule() string {
	if n.ScheduleOverride != "" {
		return n.ScheduleOverride
	}
	return "0 0 6 * * *"
}
