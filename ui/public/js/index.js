import Vue from 'vue';

var WebSocketUrl = 'ws://localhost:9091/ws'

function StartWebSocket() {
  var soc = new WebSocket(WebSocketUrl)

  soc.onopen = function(ev) {
    soc.send(JSON.stringify({
      action: 'init',
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
    console.log('message WS', ev)
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

