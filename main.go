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

// ClientStream represents connected with websocket client
// this struct hold all connected information, like current query, interval
// last ping time, configs etc.
type ClientSession struct {
	Id int64

	Query        ClientQuery
	QueryRunners []*chan int

	Active        bool
	FetchInterval int16

	CreatedAt       time.Time
	ClosedAt        time.Time
	LastKeepaliveAt time.Time
}

type ClientQuery struct {
	RawQuery string
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

func (cs *ClientSession) Close() {
	cs.Active = false
	cs.ClosedAt = time.Now()

	for _, r := range cs.QueryRunners {
		close(*r)
	}
}

func (cs *ClientSession) StartQueryRunner(app *App, results chan struct{}, ctrl chan int) error {
	log.Printf("StartQuery runned  %+v", cs)

	go func() {
		for {
			select {
			case _, ok := <-ctrl:
				if !ok {
					log.Printf("StartQuery closed -> %+v", cs)
					return
				}
			default:
				log.Printf("StartQuery FETCH  %+v", cs)
				time.Sleep((time.Duration)(FetchIntervals[0]) * time.Second)
			}
		}
	}()

	return nil
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
