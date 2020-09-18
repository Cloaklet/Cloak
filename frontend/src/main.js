import Vue from 'vue'
import VueWait from 'vue-wait'
import App from './App.vue'
import store from './store'

Vue.config.productionTip = false
Vue.use(VueWait)

new Vue({
  render: h => h(App),
  store: store,
  wait: new VueWait({
    useVuex: true,
    vuexModuleName: 'loading',
    registerComponent: false,
    registerDirective: true,
    directiveName: 'wait',
  })
}).$mount('#app')
