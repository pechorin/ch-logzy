package main

import (
	"bytes"
	"encoding/json"
	"log"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	mapstruct "github.com/mitchellh/mapstructure"
)

var (
	InitAction        = "init"
	RunQueryAction    = "run_query"
	QueryResultAction = "query_result"
)

type WebInitData struct {
	InitData map[string]interface{}
}

type RenderedWebInitData struct {
	InitData string
}

type SocketAction struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}

type SocketActionResponse struct {
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}

type SocketQueryResultsResponse struct {
	Action  string       `json:"action"`
	Results QueryResults `json:"results"`
}

// ClientStream represents connected with websocket client
// this struct hold all connected information, like current query, interval
// last ping time, configs etc.
type ClientSession struct {
	Id           int64
	QueryRunners []*chan int

	Active        bool
	FetchInterval int

	CreatedAt       time.Time
	ClosedAt        time.Time
	LastKeepaliveAt time.Time
}

type ClientQuery struct {
	Id            int32
	Query         string
	Table         string
	FetchInterval int16 `mapstructure:"fetch_interval"`
}

type ClientQueries struct {
	Queries []ClientQuery
}

type QueryResults []map[string]interface{}
type QueryResultsCh chan QueryResults

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
	wsConn, err := WS.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		app.renderError(c, err)
		return
	}
	defer wsConn.Close()

	client, err := app.CreateClientSession()
	if err != nil {
		app.renderError(c, err)
		return
	}
	defer client.Close()

	// handle results
	resultsCh := make(QueryResultsCh)
	defer close(resultsCh)

	go func() {
		for {
			select {
			case results, ok := <-resultsCh:
				if !ok {
					break
				}

				socketRes := new(SocketQueryResultsResponse)
				socketRes.Action = QueryResultAction
				socketRes.Results = results

				jsonRes, err := json.Marshal(socketRes)
				if err != nil {
					log.Printf("Invalid marshall results %s", err.Error())
					continue
				}

				if err := wsConn.WriteJSON(socketRes); err != nil {
					log.Printf("Socket write error %s", err.Error())
					continue
				}

				log.Printf("json results -> %s", jsonRes)
			}
		}
	}()

	// handle websocket messages
	for {
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("WS LOOP ABORTED -> %s", err.Error())
			break
		}

		act := new(SocketAction)
		if err := json.Unmarshal(msg, &act); err != nil {
			log.Printf("error -> %s", err.Error())
			break
		}

		switch action := act.Action; action {
		case InitAction:
			log.Printf("INIT CONNECTION %+v", act)

			tables, err := FetchClickhouseTables(app.Clickhouse)
			if err != nil {
				log.Printf("error -> %s", err.Error())
				break
			}

			resp := SocketActionResponse{Action: "init"}
			resp.Payload = make(map[string]interface{})
			resp.Payload["tables"] = tables

			if err := wsConn.WriteJSON(resp); err != nil {
				log.Printf("error -> %s", err.Error())
				break
			}

		case RunQueryAction:
			log.Printf("RUN QUERY %+v", act)

			// close current runners
			for _, rnr := range client.QueryRunners {
				close(*rnr)
			}

			// create new
			ctrlCh := make(chan int)
			runners := make([]*chan int, 0)
			runners = append(runners, &ctrlCh)

			// assign to session
			client.QueryRunners = runners

			queries := new(ClientQueries)
			if err := mapstruct.Decode(act.Payload, queries); err != nil {
				log.Printf("decode error -> %s", err.Error())
				break
			}

			log.Printf("----- decoded queires -> %+v", queries)

			for _, query := range queries.Queries {
				if valid := query.IsValid(); valid == false {
					log.Printf("invalid query -> %s :: %+v", err.Error(), query)
					break
				}

				// start query runner
				if err := client.StartQueryRunner(app, resultsCh, ctrlCh, query); err != nil {
					log.Printf("error -> %s", err.Error())
					break
				}
			}

		default:
			log.Printf("Unknown SocketAction -> %+v", act)
		}
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

func (cs *ClientSession) Close() {
	cs.Active = false
	cs.ClosedAt = time.Now()

	for _, r := range cs.QueryRunners {
		close(*r)
	}
}

func (cs *ClientSession) StartQueryRunner(app *App, resCh QueryResultsCh, ctrl chan int, query ClientQuery) error {
	log.Printf("StartQuery runned  %+v", query)

	go func() {
		for {
			select {
			case _, ok := <-ctrl:
				if !ok {
					log.Printf("StartQuery closed -> %+v", query)
					return
				}
			default:
				log.Printf("StartQuery STARTING OF FETCHING  %+v", query)

				results, err := FetchClickhouse(app.Clickhouse, query)
				if err != nil {
					log.Printf("Error while fethcing query -> %+v", query)
					return
				}

				resCh <- results

				time.Sleep((time.Duration)(query.FetchInterval) * time.Second)
			}
		}
	}()

	return nil
}

func (q *ClientQuery) IsValid() (result bool) {
	if q.FetchInterval > 0 && len(q.Table) > 0 {
		return true
	} else {
		return false
	}
}
