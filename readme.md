# feeder 📫 

Aggregates news feed updates and sends them to your email inbox.

- Supports Atom and RSS/RDF feeds.
- Supports subscribing to feed URL directly, or scanning for a `link` tag at a given URL.
- Uses Golang [html/template](https://golang.org/pkg/html/template/#pkg-overview) to customize the email body.
- Update timestamps persisted to YAML file.
- Optionally resolves relative URLs
- Optionally uses Reddit bearer token to request RSS feeds

## Usage

- Install via `go install github.com/fgeller/feeder` or download a [release](https://github.com/fgeller/feeder/releases).
- Create a [config file](https://github.com/fgeller/feeder#example-config), customizing email settings and file paths.
- Add subscribed feeds either by:
  - maintaing the [feeds config file](https://github.com/fgeller/feeder#example-feeds-config) manually, or
  - using feeder via `feeder -subscribe https://example.com/blog/`
- Run via `feeder` manually, or set up recurring execution, e.g. via `crontab -e`
- `feeder -help` output:
```
Usage of feeder:

  -config string
        Path to config file (default $XDG_CONFIG_HOME/feeder/config.yml)
  -subscribe string
        URL to feed to subscribe to
  -version
        Print version information

By default feeder will try to download the configured feeds and send
the latest entries via email. If the subscribe flag is provided, 
instead of downloading feeds, feeder tries to subscribe to the feed 
at the given URL and persists the augmented feeds config.
```

## Configuration

- `feeds-file` is the list of feeds you are subscribed to.

- `timestamp-file` is required to persist what updates have been seen.

- `email-template-file` is an optional Golang [html/template](https://golang.org/pkg/html/template/#pkg-overview) to format the sent email.

- `email` contains the configuration for sending emails. The `from` address will
  also be the `to` address and the `smtp` object allows for standard smtp host
  and auth configuration.

- `max-entries-per-feed` is the maximum number of entries to send per feed.

- `reddit` allows configuring `client-id` and `client-secret` so feeder can request and use a bearer token for Reddit RSS feeds.

### Example Config

```yaml
feeds-file: '/home/fgeller/.config/feeder/feeds.yml'
timestamp-file: '/home/fgeller/.config/feeder/timestamps.yml'
email-template-file: '/home/fgeller/.config/feeder/email.tmpl'
reddit:
  client-secret: some-secret-characters
  client-id: some-id-characters
email:
  from: example@gmail.com
  smtp:
    host: smtp.gmail.com
    port: 587
    user: example@gmail.com
    pass: passwort
```

### Example Feeds Config

```yaml
- name: 'irreal'
  url: https://irreal.org/blog/?feed=rss2
- name: The Go Blog
  url: https://blog.golang.org/blog/feed.atom
```

## Alternatives

- [blogtrottr](https://blogtrottr.com)
- [mailbrew](https://mailbrew.com/)
