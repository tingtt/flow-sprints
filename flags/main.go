package flags

import (
	"flag"
)

type AllowOrigins []string

// Implements from flag.Value
func (i *AllowOrigins) String() string {
	return "my string representation"
}

// Implements from flag.Value
func (i *AllowOrigins) Set(v string) error {
	*i = append(*i, v)
	return nil
}

type Flags struct {
	Port               *uint
	LogLevel           *uint
	GzipLevel          *uint
	AllowOrigins       AllowOrigins
	MysqlHost          *string
	MysqlPort          *uint
	MysqlDB            *string
	MysqlUser          *string
	MysqlPasswd        *string
	JwtIssuer          *string
	JwtSecret          *string
	ServiceUrlProjects *string
}

var flags Flags

func Get() Flags {
	if flags.Port == nil {
		return parse()
	}
	return flags
}

// Priority: command line params > env variables > default value
func parse() Flags {
	flags = Flags{
		flag.Uint("port", getUintEnv("PORT", 1323), "Server port"),
		flag.Uint("log-level", getUintEnv("LOG_LEVEL", 2), "Log level (1: 'DEBUG', 2: 'INFO', 3: 'WARN', 4: 'ERROR', 5: 'OFF', 6: 'PANIC', 7: 'FATAL'"),
		flag.Uint("gzip-level", getUintEnv("GZIP_LEVEL", 6), "Gzip compression level"),
		AllowOrigins{},
		flag.String("mysql-host", getEnv("MYSQL_HOST", "db"), "MySQL host"),
		flag.Uint("mysql-port", getUintEnv("MYSQL_PORT", 3306), "MySQL port"),
		flag.String("mysql-database", getEnv("MYSQL_DATABASE", "flow-sprints"), "MySQL database"),
		flag.String("mysql-user", getEnv("MYSQL_USER", "flow-sprints"), "MySQL user"),
		flag.String("mysql-password", getEnv("MYSQL_PASSWORD", ""), "MySQL password"),
		flag.String("jwt-issuer", getEnv("JWT_ISSUER", "flow-users"), "JWT issuer"),
		flag.String("jwt-secret", getEnv("JWT_SECRET", ""), "JWT secret"),
		flag.String("service-url-projects", getEnv("SERVICE_URL_PROJECTS", ""), "Service url: flow-projects"),
	}
	flag.Var(&flags.AllowOrigins, "allow-origin", "CORS allow origins")

	flag.Parse()
	return flags
}
