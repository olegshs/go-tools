# go-tools

A small toolkit for building web applications in Go. Each package can be used on its own and reads its settings from a shared configuration store:

| Package | Purpose |
|---|---|
| `config` | Global key/value configuration loaded from JSON or YAML files |
| `logs` | Leveled logging (RFC 5424 levels) through named channels |
| `events` | Event dispatcher with loosely typed listeners |
| `router` | HTTP router on top of `httprouter`: named routes, groups, prefixes, middleware, regexp constraints |
| `server` | `http.Server` wrapper with static files and access logging |
| `session` | Server-side sessions in files or an SQL table |
| `auth` | Password login and signed "remember me" cookies |
| `cache` | Cache storages: memory, files, Memcached, Redis, blackhole |
| `database` | `database/sql` wrapper with named connections and query events |
| `database/query` | Fluent SQL query builder |
| `database/orm` | Reflection-based ORM with relations and optional caching |
| `view` | `html/template` rendering with layouts and helper functions |
| `sanitize` | HTML allow-list sanitizer |
| `helpers/...` | Type conversion, struct mapping, base conversion, reflection calls, misc utilities |

## Installation

```sh
go get github.com/olegshs/go-tools@latest
```

Import packages by their path in the module, for example `github.com/olegshs/go-tools/router`. You need Go 1.26 or later, because that is what the version of `golang.org/x/net` in `go.mod` requires.

The `database` package does not register any `database/sql` driver. Import the one you need in the application:

```go
import (
	_ "github.com/go-sql-driver/mysql" // driver: mysql
	_ "github.com/lib/pq"              // driver: postgres
	_ "github.com/mattn/go-sqlite3"    // driver: sqlite3 (needs CGO)
)
```

## Quick start

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/olegshs/go-tools/config"
	"github.com/olegshs/go-tools/events"
	"github.com/olegshs/go-tools/logs"
	"github.com/olegshs/go-tools/router"
	"github.com/olegshs/go-tools/server"
)

func main() {
	// config/server.yaml -> section "server", config/logs.yaml -> "logs", ...
	if err := config.LoadAll("config/*.yaml"); err != nil {
		panic(err)
	}

	// Flush logs and release resources on Ctrl+C or SIGTERM.
	events.DispatchExitOnInterrupt()

	router.Get("/hello/{name}").Name("hello").HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		name := router.ParamsFromRequest(r).ByName("name")
		fmt.Fprintf(w, "Hello, %s!\n", name)
	})

	logs.Error(server.Start(router.DefaultRouter()))
}
```

## Configuration

`config.Load(section, file)` loads one file into a section, and `config.LoadAll(pattern)` loads every matching file, using the file name as the section name. Files ending in `.json` are parsed as JSON and everything else as YAML. If a file with the same name exists in a `local/` subdirectory next to it (e.g. `config/local/database.yaml`), it is merged on top. This is useful for machine-specific settings that should not be committed.

Values can be read with dotted keys:

```go
config.GetString("database.default.driver", "sqlite3")
config.GetInt("server.max_header_bytes", 0)
```

Packages read their whole section into a struct that starts out holding the defaults, so a file only needs the values that differ. Keys match a field's `json` tag. Durations can be written as Go duration strings (`"30s"`, `"1h30m"`) or as numbers of seconds. Relative paths are resolved against `config.BaseDir`, which is the working directory at startup.

### Settings and defaults

Settings for named instances (`logs.<name>`, `cache.<name>`, `database.<name>`) sit under the instance name. Log channels and cache storages without their own section use the `default` one.

```yaml
# server.yaml
listen: ":8080"
static: { enabled: true, dir: public }   # serve existing files from public/
log: { enabled: true, channel: default }
access_log: { enabled: false, channel: default }
tokens: true                             # X-Powered-By: go-tools
read_timeout: 0                          # also read_header_timeout, write_timeout,
                                         # idle_timeout, max_header_bytes

# logs.yaml
default:
  driver: file              # the only driver
  path: /dev/stdout         # a file path, /dev/stdout, /dev/stderr or /dev/null
  permissions: "0640"
  level: debug              # emergency alert critical error warning notice info debug
  milliseconds: false
  color: false
  forward: []               # other channels that receive a copy of every message

# cache.yaml
default:
  driver: memory            # memory (default), file, memcached, redis, blackhole
  ttl: { default: 600, maximum: 0 }
  log: { enabled: false, channel: default, interval: 60 }
  # memory:    size: 32 (MiB, 0 = unlimited), gc: { interval: 60 }
  # file:      dir: tmp/cache, hash: md5, gc: { interval: 0 }
  # memcached: servers: [{ host: localhost, port: 11211 }], prefix: ""
  # redis:     host: localhost, port: 6379, db: 0, password: "", prefix: "",
  #            pool: { max_idle, max_active, idle_timeout, wait, max_conn_lifetime }

# database.yaml
default:
  driver: postgres          # mysql, postgres or sqlite3
  host: localhost
  port: 5432                # 3306 for mysql
  username: test
  password: test
  database: test
  params: { sslmode: disable }   # mysql: { charset: utf8, parseTime: true }
  # sqlite3 uses `file: test.db` instead of host/port/credentials
  log: { enabled: false, channel: default }

# session.yaml
id: { length: 16 }          # random bytes; the cookie value is twice as long
cookie: { name: sess_id, path: /, domain: "", secure: false }
ttl: 0                      # 0 = sessions never expire
storage:
  driver: file              # file or sql
  dir: tmp/session          # file storage
  # sql storage: database: default, table: sessions, encoding: base64
  gc: { interval: 0 }

# auth.yaml
password: { hash: sha256, salt: "" }      # hash: any helpers.HashByName name or "none"
cookie:
  name: auth
  path: /
  ttl: 168h
  hash: sha256
  secret: ""                # set a long random value; it signs remember-me cookies
session: { user_id_key: auth.user.id }
```

## Usage

### Logging

```go
logs.Info("user", id, "logged in")              // default channel
logs.Channel("db").Error("query failed:", err)  // named channel
```

### Routing

Path parameters are written as `{name}`. A catch-all parameter is written as `{name...}`. Several routes can share a path shape and differ only in their regexp constraints; the first one that matches wins.

```go
r := router.New()
r.Use(loggingMiddleware)

r.Prefix("/api", func(api *router.Router) {
	api.Use(authMiddleware)
	api.Get("/articles/{id}").
		Where("id", regexp.MustCompile(`^\d+$`)).
		Name("articles.get").
		HandleFunc(getArticle)
})

url, _ := r.Url("articles.get", 42) // "/api/articles/42"
```

Routes can also be described by a map, for example one loaded from config. See `Router.ParseMap` and `ExampleRouter_ParseMap` in `router/router_test.go`.

### Database and query builder

```go
db, err := database.Get("default")
if err != nil {
	return err
}

rows, err := db.Select("p.id", "p.title", query.Expr("COUNT(c.id)")).
	From("posts p").
	LeftJoin("comments c", "c.post_id = p.id").
	Where(query.Eq{"p.status": 1}, query.Gt{"p.created": since}).
	Group("p.id", "p.title").
	Order(query.Desc("p.created")).
	Limit(20).
	Rows()
if err != nil {
	return err
}
defer rows.Close()

_, err = db.Insert("posts", query.Data{"title": "Hello", "status": 1}).Exec()
```

Conditions are combined with `AND`. Use `query.Eq`, `Ne`, `Lt`, `Lte`, `Gt`, `Gte`, `Like`, `NotLike`, `In`, `NotIn` and `Between` (maps of column to value), group them with `And`, `Or` and `Not`, or write raw SQL with `query.Expr`. Values are always sent as query arguments. A `nil` value produces `IS NULL`.

Every query dispatches an event on `db.Events()` (`database.EventQuery`, `EventExec`, ...) with its SQL, arguments, timing and error. Setting `log.enabled: true` on the connection logs every query this way.

### ORM

```go
type Model struct {
	Id int64 `orm:"primary; auto_increment"`
}

type User struct {
	Model
	Email string
	Posts []*Post `orm:"order=id desc"` // has many, via posts.user_id
}

type Post struct {
	Model
	UserId  int64
	Title   string
	Created *time.Time
	User    *User // belongs to, via user_id
}

post := &Post{UserId: 1, Title: "Hello"}
err := orm.Create(post) // sets post.Id

var posts []*Post
err = orm.Where(query.Eq{"user_id": 1}).With("User").Find(&posts)

user := &User{Model: Model{Id: 1}}
err = orm.With("Posts").First(user)
```

Tables are the snake_case plural of the type name (`users`, `posts`), and columns are snake_case field names. See the `database/orm` package documentation for all `orm` tag options. To cache loaded models, call `orm.SetCacheStorageByName("default")`.

### Sessions and authentication

```go
auth.UserFactory = myUserFactory{} // implements auth.UserFactoryInterface

func login(w http.ResponseWriter, r *http.Request) {
	s := session.Start(r, w)
	defer s.Close()

	if auth.Login(s, w, r.FormValue("login"), r.FormValue("password")) {
		if r.FormValue("remember") != "" {
			auth.Remember(s, r, w, false)
		}
	}
}

func currentUser(w http.ResponseWriter, r *http.Request, s session.SessionInterface) auth.UserInterface {
	if user := auth.User(s); user != nil {
		return user
	}
	return auth.Restore(s, r, w) // log in from the remember-me cookie
}
```

Session values are encoded with `encoding/gob`, so custom types must be registered with `gob.Register`. The SQL session storage expects a table with the columns `id` (string, primary key), `time` (integer, Unix milliseconds) and `data` (text, or binary with `encoding: none`).

### Cache

```go
var stats Stats
if err := cache.Get("stats", &stats); err != nil {
	stats = computeStats()
	cache.Set("stats", stats, 5*time.Minute)
}

raw := cache.Storage("pages") // a named storage working with []byte
```

Pass `cache.DefaultTTL` to use the configured default TTL, or `cache.MaximumTTL` to use the configured maximum.

### Views

```go
v := view.New()
v.SetLayout("views/layout.html")
html, err := v.Render("views/index.html", map[string]interface{}{"Title": "Home"})
```

Inside templates, use `{{layout "..."}}`, `{{parent "..."}}`, `{{render "..." args}}` and `{{url "route.name" params}}`. The layout or parent receives the rendered content as `{{.Content}}`. See the `view` package documentation for the full list of helper functions.

### HTML sanitizing

```go
clean := sanitize.Html(input, []string{"p", "b", "i", "a[href,title]"}, sanitize.HtmlOptions{})
```

Tags that are not allowed are removed but their text is kept. `javascript:` links in `<a href>` are removed unless `JsHrefAllow` is set.

## Development

```sh
go build ./...
go vet ./...
go test ./...                               # all tests (sqlite tests need CGO)
go test ./router -run TestRouter_Url -v     # a single test
```

## Known limitations

- Remember-me tokens are signed with a plain hash of the payload and secrets, not an HMAC.
- Passwords are stored as a single fast hash (SHA-256 by default), not with a password hashing function such as bcrypt or Argon2.

Changing either one would invalidate existing remember-me cookies or stored passwords.
