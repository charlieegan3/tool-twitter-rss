package tool

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/charlieegan3/toolbelt/pkg/apis"
	"github.com/gorilla/mux"
)

// TwitterRSS is a tool for creating RSS feeds from run/bike/swim etc activities found on a certain website
type TwitterRSS struct {
	config *gabs.Container
	db     *sql.DB
}

func (t *TwitterRSS) Name() string {
	return "twitter-rss"
}

func (t *TwitterRSS) FeatureSet() apis.FeatureSet {
	return apis.FeatureSet{
		Config: true,
		Jobs:   true,
	}
}

func (t *TwitterRSS) SetConfig(config map[string]any) error {
	t.config = gabs.Wrap(config)

	return nil
}
func (t *TwitterRSS) Jobs() ([]apis.Job, error) {
	var j []apis.Job
	var path string
	var ok bool

	// load all config
	path = "jobs.new-entry.schedule"
	schedule, ok := t.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	path = "jobs.new-entry.endpoint"
	endpoint, ok := t.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	path = "jobs.new-entry.twitter.access_token"
	accessToken, ok := t.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	path = "jobs.new-entry.twitter.access_token_secret"
	accessTokenSecret, ok := t.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	path = "jobs.new-entry.twitter.consumer_key"
	consumerKey, ok := t.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}
	path = "jobs.new-entry.twitter.consumer_secret"
	consumerSecret, ok := t.config.Path(path).Data().(string)
	if !ok {
		return j, fmt.Errorf("missing required config path: %s", path)
	}

	return []apis.Job{
		&NewEntry{
			ScheduleOverride:  schedule,
			Endpoint:          endpoint,
			AccessToken:       accessToken,
			AccessTokenSecret: accessTokenSecret,
			ConsumerKey:       consumerKey,
			ConsumerSecret:    consumerSecret,
		},
	}, nil
}
func (t *TwitterRSS) ExternalJobsFuncSet(f func(job apis.ExternalJob) error) {}

func (t *TwitterRSS) DatabaseMigrations() (*embed.FS, string, error) {
	return &embed.FS{}, "migrations", nil
}
func (t *TwitterRSS) DatabaseSet(db *sql.DB)              {}
func (t *TwitterRSS) HTTPPath() string                    { return "" }
func (t *TwitterRSS) HTTPAttach(router *mux.Router) error { return nil }
