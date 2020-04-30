<template>
  <v-app>

    <v-app-bar
      app
      clippedLeft
      class="AppBar"
    >
      <div class="d-flex align-center">
        <v-app-bar-nav-icon @click="drawer = !drawer" color="#f8bc25" />
        <h1 class="headline font-weight-black AppTitle">ch-logzy ~</h1>
      </div>
    </v-app-bar>

    <v-navigation-drawer
      v-model="drawer"
      app
      clipped
      color="grey lighten-4"
      class="MainMenu"
    >
      <v-list-item>
        <v-list-item-content>
          <v-list-item-title class="subtitle-1">
            Стримы 🌊
          </v-list-item-title>
          <v-list-item-action-text>
            Помогут следить за интересующими событиями в разных таблицах базы данных
          </v-list-item-action-text>
        </v-list-item-content>
      </v-list-item>

      <v-divider class="StreamItemDivider"></v-divider>

      <v-list-item link class="ListItem">
        <v-list-item-content>
          <v-list-item-title class="subtitle-1 StreamItemTitle">
            Все логи
          </v-list-item-title>
          <v-list-item-action-text class="StreamItemText">
            ch_logs_test_data
          </v-list-item-action-text>
        </v-list-item-content>
      </v-list-item>

      <v-divider class="StreamItemDivider"></v-divider>

      <v-list-item link class="ListItem">
        <v-list-item-content>
          <v-list-item-title class="subtitle-1 StreamItemTitle">
            Все события
          </v-list-item-title>
          <v-list-item-action-text class="StreamItemText">
            dredd_development
          </v-list-item-action-text>
        </v-list-item-content>
      </v-list-item>

      <v-divider class="StreamItemDivider"></v-divider>

      <v-list-item link class="ListItem">
        <v-list-item-content>
          <v-list-item-title class="subtitle-1 StreamItemTitle">
            Логи воркеров
          </v-list-item-title>
          <v-list-item-action-text class="StreamItemText">
            dredd_development
          </v-list-item-action-text>
        </v-list-item-content>
      </v-list-item>

      <v-divider class="StreamItemDivider"></v-divider>

      <v-list-item link class="AddButton">
        <v-list-item-content>
          <v-list-item-title class="subtitle-2">
            <v-btn class="mx-1" fab x-small depressed color="#f5853a">
              <v-icon>mdi-plus</v-icon>
            </v-btn>
            Открыть стрим
          </v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-divider></v-divider>

    </v-navigation-drawer>

    <v-content>
      <router-view />
    </v-content>
  </v-app>
</template>

<style>
.v-navigation-drawer__border {
  background-color: hsl(306, 11%, 73%);
}

.AppBar {
  background: rgb(111, 114, 150) !important;
  background: linear-gradient(90deg, rgb(111, 114, 150) 0%, rgb(158, 121, 162) 35%, rgb(129, 117, 155) 100%) !important
}

.AppBar .AppTitle {
  color: #FFFFFF;
}

.MainMenu .ListItem {
  background-color: #f7fcff;
}

.MainMenu .AddButton {
  background-color: #ffefbf;
}

.MainMeny .StreamItemTitle {
  color: #7a0786;
}
.StreamItemText {

}

.MainMenu .StreamItemDivider {
  border-color: #e8e8e8 !important;
}
</style>

<script>
export default {
  name: 'app',
  data: () => ({
    drawer: true
  }),
};
</script>

