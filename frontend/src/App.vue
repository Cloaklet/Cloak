<script setup lang="ts">
import '@vianetz/spectre.css/dist/spectre.min.css'
import 'remixicon/fonts/remixicon.css'
import Alert from './components/Alert.vue'
import VaultList from './components/VaultList.vue'
import AppHeader from "./components/AppHeader.vue"
import VaultInfoPanel from "./components/VaultInfoPanel.vue"

</script>
<template>
  <transition appear name="fade">
    <div id="cloak" class="container hero">
      <div class="columns col-gapless">
        <div class="column col-mx-auto card"><!--Limit width-->
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


<style>
#cloak {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  border-radius: 0;
  width: 650px;
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

