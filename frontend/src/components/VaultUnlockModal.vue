<template>
  <div class="modal active" @keydown.esc="close">
    <a class="modal-overlay" aria-label="Close" @click="close"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="close"></a>
        <div class="modal-title h5">{{ selectedVault.name }}</div>
      </div>
      <div class="modal-body">
        <div class="content">
          <div class="form-group">
            <i18n tag="label" for="vault-password" class="form-label" path="panel.unlock.password.label">
              <template #vaultname>{{ selectedVault.name }}</template>
            </i18n>
            <input type="password"
                   class="form-input"
                   id="vault-password"
                   ref="passwordInput"
                   v-model="password"
                   @keydown.enter="requestUnlockVault">
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary"
                :disabled="!password.length"
                :class="{ loading: $wait.is('unlocking') }"
                v-wait:disabled="'unlocking'"
                v-wait:click.start="'unlocking'"
                @click="requestUnlockVault"
                v-t="'misc.unlock'"></button>
        <button class="btn ml-1" aria-label="Close" @click="close" v-t="'misc.cancel'"></button>
      </div>
    </div>
  </div>
</template>

<script>
import {mapGetters} from "vuex";

export default {
  name: "VaultUnlockModal",
  data: function () {
    return {
      password: ''
    }
  },
  computed: {
    ...mapGetters(['selectedVault'])
  },
  methods: {
    close() {
      this.$emit('close')
    },
    requestUnlockVault() {
      this.$emit('unlock-vault-request', {
        vaultId: this.selectedVault.id,
        password: this.password
      })
    }
  },
  mounted() {
    this.$nextTick(() => this.$refs.passwordInput.focus())
  }
}
</script>

<style scoped>

</style>