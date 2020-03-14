package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/namsral/flag"
)

// App represents the application with configuration state
type App struct {
	Port          string
	AssetsDir     string
	AssetsBox     *packr.Box
	ClickhouseUri string // move to engine abstraction?

	Debug bool
}

// New creates new application instance
func New() *App {
	app := new(App)
	app.Port = ":9091"
	app.AssetsDir = "./ui/dist"
	app.AssetsBox = packr.New("Assets", app.AssetsDir)

	flag.StringVar(&app.ClickhouseUri, "clickhouse_uri", "http://localhost:8123", "Clickhouse uri with scheme")
	flag.BoolVar(&app.Debug, "debug", true, "debug output")
	flag.Parse()

	if app.Debug {
		fmt.Printf("Initial config -> %v", app)
	}

	return app
}

func main() {
	app := New()
	router := gin.Default()

	router.Static("/js", app.AssetsDir+"/js")
	router.Static("/css", app.AssetsDir+"/css")
	router.Static("/img", app.AssetsDir+"/img")
	router.Static("/fonts", app.AssetsDir+"/fonts")

	router.GET("/", app.handleIndex)
	router.GET("/about", app.handleIndex)
	router.GET("/favicon.ico", app.handleFavicon)
	router.Run(app.Port)
}

func (app *App) renderError(c *gin.Context, err error) {
	c.String(500, err.Error())
}

func (app *App) handleIndex(c *gin.Context) {
	html, err := app.AssetsBox.FindString("index.html")

	if err != nil {
		app.renderError(c, err)
		return
	}

	c.Writer.WriteHeader(200)
	c.Writer.WriteString(html)
}

func (app *App) handleFavicon(c *gin.Context) {
	html, err := app.AssetsBox.FindString("favicon.ico")

	if err != nil {
		app.renderError(c, err)
		return
	}

	c.Writer.WriteHeader(200)
	c.Writer.WriteString(html)
}
