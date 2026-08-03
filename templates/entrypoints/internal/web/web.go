// Package web implements the entrypoint flow shared by the web frontend
// images: PHP and web server configuration on top of a prepared database.
package web

import (
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/bootstrap"
	"github.com/zabbix/zabbix-docker/templates/entrypoints/internal/hooks"
)

// DB abstracts the frontend database backend, implemented by the
// mysql and postgresql packages.
type DB interface {
	Configure(string) error
	Wait() error
	Name() string
	Schema() string
	User() string
	Password() string
}

// DBType selects the database flavour of the frontend image.
type DBType string

const (
	MySQL      DBType = "MYSQL"
	PostgreSQL DBType = "POSTGRESQL"
)

// ServerType selects the web server of the frontend image.
type ServerType string

const (
	Apache ServerType = "apache"
	Nginx  ServerType = "nginx"
)

const dbName = "zabbix"

// Options describes the concrete frontend image flavour.
type Options struct {
	DBType DBType
	Server ServerType
}

// Run prepares PHP, the web server and the database connection, executes
// the entrypoint hooks and finally starts the configured server stack.
// Non-empty args short-circuit the preparation and are executed instead.
func Run(env bootstrap.Environment, db DB, opts Options, args []string) error {
	if len(args) != 0 {
		return bootstrap.Execute(args, env)
	}

	prepareWebServer := prepareNginx

	env.SetDefaultNonEmpty("ZBX_SERVER_NAME", "Zabbix docker")
	env.SetDefaultNonEmpty("PHP_TZ", "Europe/Riga")

	if opts.Server == Apache {
		env.SetDefaultNonEmpty("DAEMON_USER", "apache")
		env.SetDefaultNonEmpty("DAEMON_GROUP", "apache")

		prepareWebServer = prepareApache
	} else {
		env.SetDefaultNonEmpty("DAEMON_USER", "nginx")
		env.SetDefaultNonEmpty("DAEMON_GROUP", "nginx")
	}

	if err := db.Configure(dbName); err != nil {
		return err
	}

	if err := db.Wait(); err != nil {
		return err
	}

	if err := preparePHP(env, opts.DBType); err != nil {
		return err
	}

	if err := prepareWebServer(env); err != nil {
		return err
	}

	setDBEnv(env, db)

	if err := hooks.Run(env); err != nil {
		return err
	}

	clearWebEnv(env)

	return startStack(env, opts.Server)
}

func setDBEnv(env bootstrap.Environment, db DB) {
	env["DB_SERVER_DBNAME"] = db.Name()
	env["DB_SERVER_SCHEMA"] = db.Schema()
	env["DB_SERVER_USER"] = db.User()
	env["DB_SERVER_PASS"] = db.Password()
}

func clearWebEnv(env bootstrap.Environment) {
	prefixes := []string{"MYSQL_", "POSTGRES_"}

	if env["ZBX_VAULT"] != "" {
		prefixes = append(prefixes, "DB_SERVER_USER", "DB_SERVER_PASS")
	}

	bootstrap.ClearPrivateEnv(env, prefixes...)
}
