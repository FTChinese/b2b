package main

import (
	"flag"
	"fmt"
	"github.com/FTChinese/b2b/controllers"
	"github.com/FTChinese/b2b/database"
	"github.com/FTChinese/b2b/repository"
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

var (
	isProduction bool
	version      string
	build        string
	config       Config
	logger       = logrus.WithField("project", "superyard").WithField("package", "main")
)

func init() {
	flag.BoolVar(&isProduction, "production", false, "Indicate productions environment if present")
	var v = flag.Bool("v", false, "print current version")

	flag.Parse()

	if *v {
		fmt.Printf("%s\nBuild at %s\n", version, build)
		os.Exit(0)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		os.Exit(1)
	}

	config = Config{
		Debug:   !isProduction,
		Version: version,
		BuiltAt: build,
		Year:    0,
	}
}

func main() {
	db, err := database.New(config.MustGetDBConn("mysql.master"))
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	emailConn := MustGetEmailConn()
	post := postoffice.New(
		emailConn.Host,
		emailConn.Port,
		emailConn.User,
		emailConn.Pass)

	repo := repository.NewEnv(db)

	barrierRouter := controllers.NewBarrierRouter(repo, post)

	e := echo.New()
	e.Pre(middleware.AddTrailingSlash())

	e.Renderer = MustNewRenderer(config)
	e.HTTPErrorHandler = errorHandler

	if !isProduction {
		e.Static("/css", "client/node_modules/bootstrap/dist/css")
		e.Static("/js", "client/node_modules/bootstrap.native/dist")
		e.Static("/static", "build/dev")
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Use(middleware.CSRF())

	e.GET("/", func(context echo.Context) error {
		return context.Render(http.StatusOK, "home.html", nil)
	}, controllers.RequireLoggedIn)

	api := e.Group("/api")
	api.POST("/login/", barrierRouter.Login)
	api.POST("/signup/", barrierRouter.SignUp)

	pwResetGroup := api.Group("/password-reset")
	{
		// Handle resetting password
		pwResetGroup.POST("/", barrierRouter.ResetPassword)

		// Sending forgot-password email
		pwResetGroup.POST("/letter/", barrierRouter.PasswordResetEmail)

		// Verify forgot-password token.
		// If valid, redirect to /forgot-password.
		// If invalid, redirect to /forgot-password/letter to ask
		// user to enter email again.
		pwResetGroup.GET("/token/:token/", barrierRouter.VerifyPasswordToken)
	}

	e.Logger.Fatal(e.Start(":3100"))
}
