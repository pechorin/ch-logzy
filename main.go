package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/gorilla/websocket"
	"github.com/namsral/flag"
)

var (
	// Time intervals for fetching results from clickhouse
	FetchIntervals = [...]int16{5, 10, 15, 30, 60, 120, 240}

	// Websocket engine
	WS = websocket.Upgrader{}
)

// App represents the application with configuration state
type App struct {
	Port          string
	Debug         bool
	AssetsDir     string
	AssetsBox     *packr.Box
	IndexTemplate *template.Template

	Clickhouse    *sqlx.DB
	ClickhouseUri string // move to engine abstraction?

	Clients            []*ClientSession
	ClientsIdSerial    int64
	ClientsIdSerialMux *sync.Mutex
}

func (app *App) Log(message string) {
	if app.Debug {
		fmt.Println(message)
	}
}

// Create new application instance
func New() *App {
	app := new(App)

	clk, err := NewClickhouse()

	if err != nil {
		log.Fatalf("can't connect to clickhouse: %s", err.Error())
	}

	if err := clk.Ping(); err != nil {
		log.Fatalf("can't ping clickhouse: %s", err.Error())
	}

	app.Clickhouse = clk
	app.Port = ":9091"
	app.AssetsDir = "./ui/dist"
	app.AssetsBox = packr.New("Assets", app.AssetsDir)
	app.IndexTemplate = app.RenderIndexTemplate()
	app.ClientsIdSerialMux = new(sync.Mutex)
	app.Clients = make([]*ClientSession, 0)

	flag.StringVar(&app.ClickhouseUri, "clickhouse_uri", "http://localhost:8123", "Clickhouse uri with scheme")
	flag.BoolVar(&app.Debug, "debug", true, "debug output")
	flag.Parse()

	app.Log(fmt.Sprintf("initial config -> %+v", app))

	return app
}

// move to monitor log ?
func (app *App) ClientSessionsMonitor() error {
	timer := time.Tick(30 * time.Second)

	go func() {
		for {
			select {
			case <-timer:
				log.Printf("MONITOR - sessions: %d, fetching: %d", len(app.Clients), 0)
			}
		}
	}()

	return nil
}

func (app *App) CreateClientSession() (cs *ClientSession, err error) {
	cs = new(ClientSession)

	// increase global id
	app.ClientsIdSerialMux.Lock()
	app.ClientsIdSerial += 1
	cs.Id = app.ClientsIdSerial
	app.ClientsIdSerialMux.Unlock()

	// set defaults
	cs.Active = false
	cs.CreatedAt = time.Now()
	cs.QueryRunners = make([]*chan int, 0)

	// append to client pool
	app.Clients = append(app.Clients, cs)

	return cs, err
}

func main() {
	app := New()
	defer app.Clickhouse.Close()

	go app.ClientSessionsMonitor()

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
