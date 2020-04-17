package main

import (
	"bytes"
	"encoding/json"
	"log"
	"text/template"

	"github.com/gin-gonic/gin"
)

var (
	InitAction     = "init"
	RunQueryAction = "run_query"
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

			tables, err := AvailableTables(app.Clickhouse)
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

			resultsCh := make(chan struct{})

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

			// start new query runners
			if err := client.StartQueryRunner(app, resultsCh, ctrlCh); err != nil {
				log.Printf("error -> %s", err.Error())
				break
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
