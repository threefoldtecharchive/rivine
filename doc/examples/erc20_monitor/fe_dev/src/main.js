import Vue from "vue";
import './plugins/axios'
import App from "./App.vue";

Vue.config.productionTip = false;

new Vue({
  render: function(h) {
    return h(App);
  }
}).$mount("#app");
