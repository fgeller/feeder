# feeder 📫 

Aggregates news feed updates and sends them to your email inbox.

- Supports Atom and RSS feeds.
- Uses Golang [html/template](https://golang.org/pkg/html/template/#pkg-overview) to customize email body.
- Update timestamps persisted to YAML file.

## Usage

- Install via `go install github.com/fgeller/feeder` or download a [release](https://github.com/fgeller/feeder/releases).
- Create [configuration file](https://github.com/fgeller/feeder#configuration), customizing email settings and feeds.
- Run via `./feeder -config your-config.yml`
- Set up recurring execution, e.g. via `crontab -e`

## Configuration

- `timestamp-file` is required to persist what updates have been seen.

- `email-template-file` is an optional Golang [html/template](https://golang.org/pkg/html/template/#pkg-overview) to format the sent email.

- `email` contains the configuration for sending emails. The `from` address will
  also be the `to` address and the `smtp` object allows for standard smtp host
  and auth configuration.

- `feeds` is an array of objects with `name` and `url` string fields

### Example:

```yaml
timestamp-file: '/Users/fgeller/.config/feeder/timestamps.yml'
email-template-file: '/Users/fgeller/.config/feeder/email.tmpl'
email:
  from: example@gmail.com
  smtp:
    host: smtp.gmail.com
    port: 587
    user: example@gmail.com
    pass: passwort

feeds:
  - name: 'irreal'
    url: https://irreal.org/blog/?feed=rss2
  - name: The Go Blog
    url: https://blog.golang.org/feed.atom
```
