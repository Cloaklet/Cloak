<template>
  <div class="card-header p-0">
    <header class="navbar bg-primary text-light">
      <section class="navbar-section">
        <a class="navbar-brand text-bold text-light text-uppercase">Cloak</a>
      </section>
      <section class="navbar-section">
        <a class="btn btn-link text-light" @click="showConfigModal = true">
          <i class="ri-settings-3-fill ri-xl"></i>
        </a>
      </section>
    </header>
    <AppConfigModal v-if="showConfigModal" @close="showConfigModal = false"/>
  </div>
</template>

<script>
import AppConfigModal from "@/components/AppConfigModal";
export default {
  name: "AppHeader",
  components: {AppConfigModal},
  data: function() {
    return {
      showConfigModal: false
    }
  }
}
</script>

<style scoped>
header {
  padding: .7rem;
}
i {
  vertical-align: middle;
}
</style>