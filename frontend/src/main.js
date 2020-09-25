import Vue from 'vue'
import VueClipboard from 'vue-clipboard2'
import VueWait from 'vue-wait'
import App from './App.vue'
import store from './store'
import i18n from './i18n'

Vue.config.productionTip = false
Vue.use(VueWait)
Vue.use(VueClipboard)

new Vue({
  render: h => h(App),
  store,
  i18n,

  wait: new VueWait({
    useVuex: true,
    vuexModuleName: 'loading',
    registerComponent: false,
    registerDirective: true,
    directiveName: 'wait',
  })
}).$mount('#app')
