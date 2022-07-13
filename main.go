package main

import (
	"flow-sprints/flags"
	"flow-sprints/handler"
	"flow-sprints/jwt"
	"flow-sprints/mysql"
	"flow-sprints/sprint"
	"flow-sprints/utils"
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	// Register custum validations
	cv.validator.RegisterValidation("Y-M-D", sprint.DateStrValidation)

	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func logFormat() string {
	// Refer to https://github.com/tkuchiki/alp
	var format string
	format += "time:${time_rfc3339}\t"
	format += "host:${remote_ip}\t"
	format += "forwardedfor:${header:x-forwarded-for}\t"
	format += "req:-\t"
	format += "status:${status}\t"
	format += "method:${method}\t"
	format += "uri:${uri}\t"
	format += "size:${bytes_out}\t"
	format += "referer:${referer}\t"
	format += "ua:${user_agent}\t"
	format += "reqtime_ns:${latency}\t"
	format += "cache:-\t"
	format += "runtime:-\t"
	format += "apptime:-\t"
	format += "vhost:${host}\t"
	format += "reqtime_human:${latency_human}\t"
	format += "x-request-id:${id}\t"
	format += "host:${host}\n"
	return format
}

func main() {
	// Get command line params / env variables
	f := flags.Get()

	//
	// Setup echo and middlewares
	//

	// Echo instance
	e := echo.New()

	// Log level
	e.Logger.SetLevel(log.Lvl(*f.LogLevel))
	e.Logger.Infof("Log level %d", *f.LogLevel)

	// Gzip
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: int(*f.GzipLevel),
	}))
	e.Logger.Infof("Gzip enabled with level %d", *f.GzipLevel)

	// CORS
	if f.AllowOrigins != nil && len(f.AllowOrigins) != 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: f.AllowOrigins,
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		}))
		e.Logger.Info("CORS enabled")
		e.Logger.Debugf("CORS allow origins %s", f.AllowOrigins.String())
	}

	// JWT
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwt.JwtCustumClaims{},
		SigningKey: []byte(*f.JwtSecret),
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/-/readiness"
		},
	}))

	// Logger
	if f.LogLevel != nil && *f.LogLevel == 1 {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: logFormat(),
			Output: os.Stdout,
			Skipper: func(c echo.Context) bool {
				return c.Path() == "/-/readiness"
			},
		}))
		e.Logger.Info("Access logging with `alp`(https://github.com/tkuchiki/alp) enabled")
	}

	// Validator instance
	e.Validator = &CustomValidator{validator: validator.New()}

	//
	// Setup DB
	//

	// DB client instance
	e.Logger.Debugf("DB DSN `%s`", mysql.SetDSNTCP(*f.MysqlUser, *f.MysqlPasswd, *f.MysqlHost, int(*f.MysqlPort), *f.MysqlDB))

	// Check connection
	d, err := mysql.Open()
	if err != nil {
		e.Logger.Fatal(err)
	}
	if err = d.Ping(); err != nil {
		e.Logger.Fatal(err)
	}
	e.Logger.Info("DB connection test succeeded")

	//
	// Check health of external service
	//
	if *flags.Get().ServiceUrlProjects == "" {
		e.Logger.Fatal("`--service-url-projects` option is required")
	}
	if status, err := utils.HttpGet(*flags.Get().ServiceUrlProjects+"/-/readiness", nil); err != nil {
		e.Logger.Fatalf("failed to check health of external service `flow-projects` %s", err)
	} else if status != http.StatusOK {
		e.Logger.Fatal("failed to check health of external service `flow-projects`")
	}
	e.Logger.Debug("Check health of external service `flow-projects` succeeded")

	//
	// Routes
	//

	// Health check route
	e.GET("/-/readiness", func(c echo.Context) error {
		return c.String(http.StatusOK, "flow-sprints is Healthy.\n")
	})

	// Restricted routes
	e.GET("/", handler.GetList)
	e.POST("/", handler.Post)
	e.GET(":id", handler.Get)
	e.PATCH(":id", handler.Patch)
	e.DELETE(":id", handler.Delete)
	e.DELETE("/", handler.DeleteAll)

	//
	// Start echo
	//
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", *f.Port)))
}
