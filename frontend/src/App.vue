<template>
  <transition appear name="fade">
    <div id="app" class="container hero">
      <div class="columns col-gapless">
        <div class="column col-7 col-mx-auto card"><!--Limit width-->
          <AppHeader/>
          <div class="columns col-gapless card-body p-0">
            <VaultList/>
            <VaultInfoPanel/>
          </div>
        </div>
      </div>
      <Alert/>
    </div>
  </transition>
</template>

<script>
import 'spectre.css/dist/spectre.min.css'
import 'remixicon/fonts/remixicon.css'
import Alert from '@/components/Alert.vue'
import VaultList from '@/components/VaultList.vue'
import AppHeader from "@/components/AppHeader"
import VaultInfoPanel from "@/components/VaultInfoPanel"

export default {
  name: 'App',
  components: {
    VaultInfoPanel,
    AppHeader,
    Alert,
    VaultList
  },
}
</script>

<style>
#app {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  border-radius: 0;
}
.card {
  box-shadow: 0 .25rem 1rem rgba(48,55,66,.15);
  min-height: 500px;
}
.menu {
  box-shadow: none;
  transform: none;
  border-radius: 0;
}
[class^=ri-] {
  vertical-align: top;
}
.tab {
  user-select: none;
}
.fade-enter-active, .fade-leave-active {
  transition: opacity .6s;
}
.fade-enter, .fade-leave-to {
  opacity: 0;
}
</style>
