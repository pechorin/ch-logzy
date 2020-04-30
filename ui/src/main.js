import Vue from 'vue';
import Vuetify from 'vuetify/lib';
import VueRouter from 'vue-router';

import Application from './Application.vue';
import IndexPage from './components/IndexPage';
import NewStreamPage from './components/NewStreamPage';
import StreamPage from './components/StreamPage';
import ErrorPage from './components/ErrorPage';

Vue.use(Vuetify)
Vue.use(VueRouter)
Vue.config.productionTip = false
Vue.config.devtools = true

var vuetify = new Vuetify({})
var routes  = [
  { path: '/', component: IndexPage, name: 'index_page' },
  { path: '/new', component: NewStreamPage, name: 'new_stream_page' },
  { path: '/streams/:id', component: StreamPage, name: 'stream_page' },
  { path: '/404', component: ErrorPage, name: 'error_page', alias: '*' },
]

var router = new VueRouter({routes})

router.beforeEach((to, from, next) => {
  console.log("router -> ", to, from, next)
})

new Vue({
  vuetify,
  router,
  render: h => h(Application)
}).$mount('#app')

var INIT_ACTION           = 'init'
var RUN_QUERY_ACTION      = 'run_query'
var QUERY_RESULT_RESPONSE = 'query_result'
var WEBSOCKET_URL         = 'ws://localhost:9091/ws'


function StartWebSocket() {
  var soc = new WebSocket(WEBSOCKET_URL)

  soc.onopen = function(ev) {
    console.log('open connection WS', ev)
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

    if (msg.action == INIT_ACTION) {
      console.log("available tables -> ", data.tables)

      if (data.tables && data.tables.length > 0) {
        console.log('will fetch first table ->', data.tables[0])

        var query_id = Math.floor(Math.random() * Math.floor(100000))

        soc.send(JSON.stringify({
          action: RUN_QUERY_ACTION,
          payload: { queries: [
            { id: query_id, query: "SELECT * FROM users", fetch_interval: 30, table: data.tables[0] }
          ] }
        }))
      }

    } else if (msg.action == QUERY_RESULT_RESPONSE) {
      console.log('query result -> ', msg)
    } else {
      console.error('unkwnown WS messages ->', msg)
    }
  }
}

document.addEventListener("DOMContentLoaded", function() {
  StartWebSocket()
})

console.log('inited')

