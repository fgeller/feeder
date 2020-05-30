# feeder

Aggregates news feeds updates and send them via email.

- Supports Atom and RSS feeds.
- Golang template to format email update 
- Update timestamps persisted to YAML file

## Configuration

- `timestamp-file` is required to persist what updates have been seen.

- `email-template-file` is a optional Golang template to format the sent email.

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
