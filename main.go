package main

import (
	"flag"
	"flow-sprints/jwt"
	"flow-sprints/mysql"
	"flow-sprints/sprint"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

func getIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var intValue, err = strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Priority: command line params > env variables > default value
var (
	port               = flag.Int("port", getIntEnv("PORT", 1323), "Server port")
	logLevel           = flag.Int("log-level", getIntEnv("LOG_LEVEL", 2), "Log level (1: 'DEBUG', 2: 'INFO', 3: 'WARN', 4: 'ERROR', 5: 'OFF', 6: 'PANIC', 7: 'FATAL'")
	gzipLevel          = flag.Int("gzip-level", getIntEnv("GZIP_LEVEL", 6), "Gzip compression level")
	mysqlHost          = flag.String("mysql-host", getEnv("MYSQL_HOST", "db"), "MySQL host")
	mysqlPort          = flag.Int("mysql-port", getIntEnv("MYSQL_PORT", 3306), "MySQL port")
	mysqlDB            = flag.String("mysql-database", getEnv("MYSQL_DATABASE", "flow-sprints"), "MySQL database")
	mysqlUser          = flag.String("mysql-user", getEnv("MYSQL_USER", "flow-sprints"), "MySQL user")
	mysqlPasswd        = flag.String("mysql-password", getEnv("MYSQL_PASSWORD", ""), "MySQL password")
	jwtIssuer          = flag.String("jwt-issuer", getEnv("JWT_ISSUER", "flow-users"), "JWT issuer")
	jwtSecret          = flag.String("jwt-secret", getEnv("JWT_SECRET", ""), "JWT secret")
	serviceUrlProjects = flag.String("service-url-projects", getEnv("SERVICE_URL_PROJECTS", ""), "Service url: flow-projects")
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	// Register custum validations
	cv.validator.RegisterValidation("Y-M-D", sprint.DateStrValidation)

	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	e := echo.New()
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: *gzipLevel,
	}))
	e.Logger.SetLevel(log.Lvl(*logLevel))
	e.Validator = &CustomValidator{validator: validator.New()}

	// Setup db client instance
	e.Logger.Info(mysql.SetDSNTCP(*mysqlUser, *mysqlPasswd, *mysqlHost, *mysqlPort, *mysqlDB))
	// Check connection
	d, err := mysql.Open()
	if err != nil {
		e.Logger.Fatal(err)
	}
	if err = d.Ping(); err != nil {
		e.Logger.Fatal(err)
	}

	// Service status check
	if *serviceUrlProjects == "" {
		e.Logger.Fatal("`--service-url-projects` option is required")
	}
	ok, err := checkHealth(*serviceUrlProjects + "/-/readiness")
	if err != nil {
		e.Logger.Fatalf("failed to check health of external service `flow-projects` %s", err)
	}
	if !ok {
		e.Logger.Fatal("failed to check health of external service `flow-projects`")
	}

	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwt.JwtCustumClaims{},
		SigningKey: []byte(*jwtSecret),
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/-/readiness"
		},
	}))

	// Health check route
	e.GET("/-/readiness", func(c echo.Context) error {
		return c.String(http.StatusOK, "flow-sprints is Healthy.\n")
	})

	// Restricted routes
	e.GET("/", getList)
	e.POST("/", post)
	e.GET(":id", get)
	e.PATCH(":id", patch)
	e.DELETE(":id", delete)
	e.DELETE("/", deleteAll)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", *port)))
}
