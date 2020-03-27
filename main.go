package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"text/template"

	"github.com/gin-gonic/gin"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/namsral/flag"
	"github.com/gorilla/websocket"
)

var (
	WS = websocket.Upgrader{}
)

// App represents the application with configuration state
type App struct {
	Port          string
	AssetsDir     string
	AssetsBox     *packr.Box
	ClickhouseUri string // move to engine abstraction?
	Clickhouse    interface{}
	IndexTemplate *template.Template

	Debug bool
}

func (app *App) Log(message string) {
	if app.Debug {
		fmt.Println(message)
	}
}

type WebInitData struct {
	InitData map[string]interface{}
}

type RenderedWebInitData struct {
	InitData string
}

// New creates new application instance
func New() *App {
	app := new(App)

	app.Port = ":9091"
	app.AssetsDir = "./ui/dist"
	app.AssetsBox = packr.New("Assets", app.AssetsDir)
	app.IndexTemplate = app.RenderIndexTemplate()

	flag.StringVar(&app.ClickhouseUri, "clickhouse_uri", "http://localhost:8123", "Clickhouse uri with scheme")
	flag.BoolVar(&app.Debug, "debug", true, "debug output")
	flag.Parse()

	conn, err := NewClickhouse()
	if err != nil {
		log.Fatalf("can't connect to clickhouse: %v", err.Error())
	}

	app.Clickhouse = conn

	app.Log(fmt.Sprintf("initial config -> %v", app))

	return app
}

func main() {
	app := New()
	router := gin.Default()

	router.Static("/js", app.AssetsDir+"/js")
	router.Static("/css", app.AssetsDir+"/css")
	router.Static("/img", app.AssetsDir+"/img")
	router.Static("/fonts", app.AssetsDir+"/fonts")

	router.GET("/", app.indexController)
	router.GET("/about", app.indexController)
	router.GET("/favicon.ico", app.faviconController)
	router.GET("/ws", app.websocketController)
	router.Run(app.Port)
}

func (app *App) renderError(c *gin.Context, err error) {
	c.String(500, err.Error())
}

func (app *App) RenderIndexTemplate() *template.Template {
	html, err := app.AssetsBox.FindString("index.html")

	if err != nil {
		panic("index.html template not found")
	}

	// move to application?
	template, err := template.New("IndexTemplate").Parse(html)

	if err != nil {
		panic("can't render index template")
	}

	return template
}

func (app *App) websocketController(c *gin.Context) {
	wsConn, err := WS.Upgrade(c.Writer, c.Req, nil)
	if err != nil {
		app.renderError(c, err)
	}
	defer wsConn.Close()

	msg := new(bytes.Buffer)

	for {
		if err := wsConn.ReadJSON(&msg); err != nil {
			log.Println("error -> %v", err.Error())
			break
		}

		log.Println("rec -> %v", msg.String())
		msg.Reset()
	}
}

func (app *App) indexController(c *gin.Context) {
	render := new(bytes.Buffer)

	init := new(WebInitData)

	init.InitData = make(map[string]interface{})
	init.InitData["debug"] = true
	init.InitData["version"] = "0.1"

	marshalled, err := json.Marshal(init.InitData)
	if err != nil {
		app.renderError(c, err)
		return
	}

	rendered := new(RenderedWebInitData)
	rendered.InitData = string(marshalled)

	if err := app.IndexTemplate.Execute(render, rendered); err != nil {
		app.renderError(c, err)
		return
	}

	c.Writer.WriteHeader(200)
	c.Writer.WriteString(render.String())
}

func (app *App) faviconController(c *gin.Context) {
	html, err := app.AssetsBox.FindString("favicon.ico")

	if err != nil {
		app.renderError(c, err)
		return
	}

	c.Writer.WriteHeader(200)
	c.Writer.WriteString(html)
}
