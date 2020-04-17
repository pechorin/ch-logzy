import Vue from 'vue';

var INIT_ACTION      = 'init'
var RUN_QUERY_ACTION = 'run_query'
var WEBSOCKET_URL    = 'ws://localhost:9091/ws'

function StartWebSocket() {
  var soc = new WebSocket(WEBSOCKET_URL)

  soc.onopen = function(ev) {
    soc.send(JSON.stringify({
      action: INIT_ACTION,
      payload: { login: '', password: '', hostString: '' }
    }))
  }

  soc.onerror = function(ev) {
    console.log('error WS', ev)
  }

  soc.onclose = function(ev) {
    console.log('closed WS', ev)
  }

  soc.onmessage = function(ev) {
    var msg  = JSON.parse(ev.data)
    var data = msg.payload
    console.log('incoming WS msg ->', msg)

    if (msg.action == INIT_ACTION) {
      console.log("available tables -> ", data.tables)

      if (data.tables && data.tables.length > 0) {
        console.log('will fetch first table ->', data.tables[0])

        soc.send(JSON.stringify({
          action: RUN_QUERY_ACTION,
          payload: { queries: [
            { query: "", fetch_interval: 30 }
          ] }
        }))
      }

    } else {
      console.error('unkwnown WS messages ->', msg)
    }
  }
}

document.addEventListener("DOMContentLoaded", function() {
  var app = new Vue({
    el: '#app',
    template: `
      <div>
        Hello world !
      </div>
    `,
  })

  StartWebSocket()

  console.log('app initied', app)
})

