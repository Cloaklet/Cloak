<template>
  <div class="column col-8 info-panel">
    <div class="empty" v-if="!selectedVault">
      <p class="empty-title h5">No vault selected</p>
      <p class="empty-subtitle">Click on a vault on the left to show its details</p>
    </div>
    <div class="bg-gray vault-info" v-else>
      <div class="tile tile-centered">
        <div class="tile-icon text-center text-light bg-primary s-circle">
          <i class="ri-lock-unlock-fill ri-lg" v-if="selectedVault.state === 'unlocked'"></i>
          <i class="ri-lock-fill ri-lg" v-else></i>
        </div>
        <div class="tile-content p-relative">
          <div class="tile-title h5">{{ selectedVault.name }}</div>
          <small class="tile-subtitle text-gray">{{ selectedVault.path }}</small>
          <span class="chip text-uppercase text-bold text-light p-absolute"
                :class="{ 'bg-primary': selectedVault.state === 'unlocked' }">{{ selectedVault.state }}</span>
        </div>
      </div><!--vault info end-->
      <div class="vault-operations text-center" v-if="selectedVault.state === 'unlocked'">
        <div class="text-center">
          <button class="btn btn-primary btn-lg"
                  @click="revealMountPointForVault({vaultId: selectedVault.id})">
            <i class="ri-folder-open-fill"></i> Reveal Drive
          </button>
        </div>
        <div class="text-center mt-2">
          <button class="btn btn-sm"
                  @click="lockVault({vaultId: selectedVault.id})">
            <i class="ri-key-2-fill"></i> Lock
          </button>
        </div>
      </div><!--vault operation buttons end-->
      <div class="vault-operations text-center" v-else>
        <button class="btn btn-primary btn-lg"
                @click="showUnlock = true">
          <i class="ri-key-2-fill"></i> Unlock...
        </button>
      </div>
    </div>
    <VaultUnlockModal v-if="showUnlock"
                      @close="showUnlock = false"
                      @unlock-vault-request="unlockVault"/>
  </div>
</template>

<script>
import {mapActions, mapGetters} from 'vuex'
import VaultUnlockModal from "@/components/VaultUnlockModal";

export default {
  name: "VaultInfoPanel",
  components: {VaultUnlockModal},
  data: function () {
    return {
      showUnlock: false
    }
  },
  computed: {
    ...mapGetters(['selectedVault'])
  },
  methods: {
    ...mapActions(['revealMountPointForVault', 'lockVault']),
    unlockVault(payload) {
      this.$store.dispatch('unlockVault', payload).then(() => {
        this.showUnlock = false
        this.$wait.end('unlocking')
      })
    }
  }
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