package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/random"
	"github.com/spf13/viper"
)

// URL struct of host to route
type URL struct {
	Host string `json:"host"`
	Port string `json:"port"`
	URI  string `json:"uri"`
}

// Version is struct for get version form conf.yaml
type Version struct {
	Version string `json:"version"`
}

var reqID string

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

func setHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gg := &reqID
		*gg = c.Request().Header.Get(echo.HeaderXRequestID)
		c.Logger().SetHeader(c.Request().Header.Get(echo.HeaderXRequestID))
		return next(c)
	}
}

func main() {
	// Echo instance
	e := echo.New()
	// Middleware
	e.Use(setHeader)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Use(middleware.RequestID())
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			if reqID == "" {
				gg := &reqID
				*gg = random.String(32)
				return reqID
			}
			return reqID
		},
	}))

	// Route => handler
	e.GET("/*", callDefault)
	e.GET("/build", callBuild)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
func setURL() string {
	url := URL{}
	url.Host = os.Getenv("HOST")
	url.Port = os.Getenv("PORT")
	url.URI = os.Getenv("URI")
	if url.Host == "" {
		return ""
	}
	urlTrim := url.Host + ":" + url.Port + url.URI
	return urlTrim
}

func callDefault(c echo.Context) error {
	url := setURL()
	if url == "" {
		fmt.Print("Can't find URL")
		return c.JSON(http.StatusInternalServerError, "Can't find URL")
	}
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("The HTTP custom new request failed with error %s\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	req.Header.Set(echo.HeaderXRequestID, reqID) // Set Header by key of echo ReqID

	respones, err := client.Do(req)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer respones.Body.Close()
	data := respones.StatusCode
	if data < 400 {
		return c.JSON(data, "Success : "+strconv.Itoa(data))
	} else if data >= 400 {
		return c.JSON(data, "Failed : "+strconv.Itoa(data))
	}
	return c.JSON(data, nil)
}

func callBuild(c echo.Context) error {
	v := Version{}
	v.Version = viper.GetString("service.version")
	return c.JSON(http.StatusOK, v)
}
