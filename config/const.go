package config

const (
	EnvPrefix string = "AVALOND"

	EnvEnvironment       string = "ENVIRONMENT"
	EnvLogLevel          string = "LOG_LEVEL"
	EnvLogKind           string = "LOG_KIND"
	EnvNatsAddress       string = "NATS_ADDRESS"
	EnvNatsEncoding      string = "NATS_ENCODING"
	EnvCouchbaseURL      string = "COUCHBASE_URL"
	EnvCouchbaseBucket   string = "COUCHBASE_BUCKET"
	EnvCouchbaseUsername string = "COUCHBASE_USERNAME"
	EnvCouchbasePassword string = "COUCHBASE_PASSWORD"

	FlagEnvironment       string = "environment"
	FlagLogLevel          string = "log-level"
	FlagLogKind           string = "log-kind"
	FlagNatsAddress       string = "nats-address"
	FlagNatsEncoding      string = "nats-encoding"
	FlagCouchbaseURL      string = "couchbase-url"
	FlagCouchbaseBucket   string = "couchbase-bucket"
	FlagCouchbaseUsername string = "couchbase-username"
	FlagCouchbasePassword string = "couchbase-password"
)
