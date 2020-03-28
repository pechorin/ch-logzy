import Vue from 'vue';

var app = new Vue({
  el: '#app',
  template: `
     <ol>
        <li v-for="item in items">
          {{ item.text }}
        </li>
      </ol>
  `,
  data: {
    items: [
      { text: 'item 1' },
      { text: 'item 2' },
      { text: 'item 3' }
    ]
  }
})

console.log('app initied', app)
