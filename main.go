package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
)

// Version is struct for get version form conf.yaml
type Version struct {
	Version string `json:"version"`
}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}

func main() {
	// Echo instance
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Route => handler
	e.GET("/*", callDefault)
	e.GET("/build", callBuild)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
func setURL() string {
	return os.Getenv("HOST") + ":" + os.Getenv("PORT") + os.Getenv("URI")
}

func callDefault(c echo.Context) error {
	url := setURL()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("The HTTP custom new request failed with error %s\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	req.Header.Set(echo.HeaderXRequestID, c.Request().Header.Get(echo.HeaderXRequestID)) // Set Header by key of echo ReqID

	respones, err := client.Do(req)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	defer respones.Body.Close()
	go cal()
	cal()

	return c.JSON(respones.StatusCode, nil)
}

func cal() {
	x := 0.0
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := 0.0; i < 100000.0; i++ {
		x = (((x + 0.24999484587428) * i) / (i * 12.4123456)) * rand.Float64()
	}
}

func callBuild(c echo.Context) error {
	v := Version{}
	v.Version = viper.GetString("service.version")
	return c.JSON(http.StatusOK, v)
}
