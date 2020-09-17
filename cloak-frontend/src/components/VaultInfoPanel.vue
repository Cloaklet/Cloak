<template>
  <div class="column col-8 info-panel">
    <div class="empty" v-if="!vault">
      <p class="empty-title h5">No vault selected</p>
      <p class="empty-subtitle">Click on a vault on the left to show its details</p>
    </div>
    <div class="bg-gray vault-info" v-else>
      <div class="tile tile-centered">
        <div class="tile-icon text-center text-light bg-primary s-circle">
          <i class="ri-lock-unlock-fill ri-lg" v-if="vault.state === 'unlocked'"></i>
          <i class="ri-lock-fill ri-lg" v-else></i>
        </div>
        <div class="tile-content p-relative">
          <div class="tile-title h5">{{ vault.name }}</div>
          <small class="tile-subtitle text-gray">{{ vault.path }}</small>
          <span class="chip text-uppercase text-bold text-light p-absolute"
                :class="{ 'bg-primary': vault.state === 'unlocked' }">{{ vault.state }}</span>
        </div>
      </div><!--vault info end-->
      <div class="vault-operations text-center" v-if="vault.state === 'unlocked'">
        <div class="text-center">
          <button class="btn btn-primary btn-lg"
                  @click="revealMountpoint(vault.id)"><i class="ri-folder-open-fill"></i> Reveal Drive</button>
        </div>
        <div class="text-center mt-2">
          <button class="btn btn-sm"
                  @click="lockVault(vault.id)"><i class="ri-key-2-fill"></i> Lock</button>
        </div>
      </div><!--vault operation buttons end-->
      <div class="vault-operations text-center" v-else>
        <button class="btn btn-primary btn-lg"
                @click="showUnlock = true"><i class="ri-key-2-fill"></i> Unlock...</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "VaultInfoPanel",
  props: ['vault']
}
</script>

<style scoped>
.info-panel {
  border-left: .05rem solid #dadee4;
}
.info-panel > div {
  height: 100%;
}
.vault-info {
  padding: 1.4rem;
}
.vault-info .chip {
  background-color: #8c8c8c;
  top: 0;
  right: 0;
  font-size: .6em;
}
.vault-info .s-circle {
  padding: .4rem;
  width: 40px;
  height: 40px;
}
.vault-info [class^=ri-] {
  vertical-align: sub;
}
.vault-operations {
  margin-top: 3rem;
}
</style>