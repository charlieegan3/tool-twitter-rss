# tool-twitter-rss

This is a tool for [toolbelt](https://github.com/charlieegan3/toolbelt) which collects tweets for the previous day and
generates an RSS feed item using [webhook-rss](https://github.com/charlieegan3/tool-webhook-rss) deployed to the same
toolbelt.

Example config:

```yaml
tools:
  ...
  twitter-rss:
    jobs:
      new-entry:
        schedule: "0 0 5 * * *"
        endpoint: https://...
        twitter:
          username: xxx
          access_token: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          access_token_secret: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          consumer_key: xxxxxxxxxxxxxxxxxxxxxx
          consumer_secret: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```
