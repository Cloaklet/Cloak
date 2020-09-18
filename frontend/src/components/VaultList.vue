<template>
  <div class="column col-4 p-relative">
    <div class="menu p-0 vault-list">
      <div class="empty" v-if="!vaultsCount">
        <p class="empty-title h5">No vault yet</p>
        <p class="empty-subtitle">Click on the add button to add an existing vault or create a new one</p>
      </div>
      <div class="menu-item mt-0" v-for="vault in $store.state.vaults" :key="vault.id">
        <a :class="{ active: vault === selectedVault }" @click="selectVault({vaultId: vault.id})">
          <div class="tile tile-centered">
            <div class="tile-icon">
              <i class="ri-lock-unlock-fill ri-lg" v-if="vault.state === 'unlocked'"></i>
              <i class="ri-lock-fill ri-lg" v-else></i>
            </div>
            <div class="tile-content">
              <div class="tile-title h5">{{ vault.name }}</div>
              <small class="tile-subtitle text-gray">{{ vault.path }}</small>
            </div>
          </div>
        </a>
      </div>
    </div><!--list end-->
    <div class="btn-group btn-group-block p-absolute vault-list-buttons">
      <button class="btn btn-lg bg-gray h6 text-normal tooltip"
              data-tooltip="Add Vault"
              @click="showAddVaultModal = true">➕</button>
      <button class="btn btn-lg bg-gray h6 text-normal tooltip"
              data-tooltip="Remove Vault"
              :disabled="!selectedVault"
              :class="{ loading: $wait.is('removing vault') }"
              v-wait:disabled="'removing vault'"
              @click="removeVault({vaultId: selectedVault.id})">➖</button>
    </div>
    <AddVaultModal v-if="showAddVaultModal"
                   @close="showAddVaultModal = false"
                   @add-vault-request="addVault"
                   @create-vault-request="createVault"/>
  </div>
</template>

<script>
import {mapGetters, mapMutations} from 'vuex'
import AddVaultModal from './AddVaultModal'
import {mapWaitingActions} from 'vue-wait'

export default {
  name: "VaultList",
  components: {
    AddVaultModal
  },
  data: function () {
    return {
      showAddVaultModal: false
    }
  },
  computed: {
    ...mapGetters(['selectedVault', 'vaultsCount'])
  },
  methods: {
    ...mapMutations(['selectVault']),
    ...mapWaitingActions({
      removeVault: 'removing vault'
    }),
    addVault(payload) {
      this.$store.dispatch('addVault', payload).then(() => {
        this.showAddVaultModal = false
        this.$wait.end('adding vault')
      })
    },
    createVault(payload) {
      this.$store.dispatch('createVault', payload).then(() => {
        this.showAddVaultModal = false
        this.$wait.end('creating vault')
      })
    }
  },
  mounted() {
    this.$store.dispatch('loadVaults')
  }
}
</script>

<style scoped>
.column {
  user-select: none;
}
.vault-list, .vault-list .empty {
  height: 100%;
}
.menu-item > a {
  border-radius: 0;
  padding: .4rem;
  border-left: 4px solid transparent;
}
.menu-item > a.active {
  border-color: #5755d9;
}
.vault-list-buttons {
  width: 100%;
  color: unset;
  bottom: 0;
}
.vault-list-buttons .btn {
  border-color: #dadee4;
  border-radius: 0;
  border-bottom: none;
}
.vault-list-buttons .btn:first-child {
  border-left: none;
}
.vault-list-buttons .btn:last-child {
  border-right: none;
}
</style>