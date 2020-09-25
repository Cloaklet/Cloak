<template>
  <div class="modal active" @keydown.esc="close">
    <a class="modal-overlay" aria-label="Close" @click="close"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="close"></a>
        <div class="modal-title h5" v-t="'vault.options.masterkey.title'"></div>
      </div>
      <div class="modal-body">
        <div class="content">
          <div class="form-group" v-if="!masterkey">
            <i18n tag="label" for="vault-password" class="form-label" path="panel.unlock.password.label">
              <template #vaultname>{{ selectedVault.name }}</template>
            </i18n>
            <input type="password"
                   class="form-input"
                   :disabled="$wait.is('revealing masterkey')"
                   id="vault-password"
                   ref="passwordInput"
                   v-model="password"
                   @keydown.enter="requestRevealMasterkey">
          </div>
          <div v-else>
            <i18n tag="p" class="form-group" path="vault.options.masterkey.description">
              <template #vaultname>{{ selectedVault.name }}</template>
            </i18n>
            <div class="form-group input-group">
              <input type="text" class="form-input" readonly v-model="masterkey">
              <button class="btn input-group-btn float-right"
                      v-clipboard:copy="masterkey"
                      v-clipboard:success="onCopySucceeded"
                      v-clipboard:error="onCopyFailed">
                <i class="ri-file-copy-fill"></i> Copy
              </button>
            </div>
            <div class="mt-2 pt-2">
              <span v-t="'vault.options.masterkey.keep_description'"></span>
              <ul class="m-0 ml-1">
                <li v-t="'vault.options.masterkey.keep_note.note_1'"></li>
                <li v-t="'vault.options.masterkey.keep_note.note_2'"></li>
              </ul>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary"
                v-if="!masterkey"
                :disabled="!password.length || $wait.is('revealing masterkey')"
                :class="{ loading: $wait.is('revealing masterkey') }"
                @click="requestRevealMasterkey"
                v-t="'vault.options.masterkey.view'"></button>
        <button class="btn ml-1"
                v-if="!masterkey"
                aria-label="Close"
                @click="close"
                v-t="'misc.cancel'"></button>
        <button class="btn ml-1" v-else @click="close" v-t="'misc.done'"></button>
      </div>
    </div>
  </div>
</template>

<script>
import {mapGetters} from 'vuex'

export default {
  name: "VaultMasterkeyModal",
  data: function() {
    return {
      password: '',
      masterkey: '',
    }
  },
  computed: {
    ...mapGetters(['selectedVault'])
  },
  methods: {
    close() {
      this.$emit('close')
    },
    requestRevealMasterkey() {
      this.$wait.start('revealing masterkey')
      this.$store.dispatch('revealVaultMasterkey', {
        vaultId: this.selectedVault.id,
        password: this.password
      }).then(masterkey => {
        this.masterkey = masterkey
      }).finally(() => {
        this.$wait.end('revealing masterkey')
      })
    },
    onCopySucceeded() {
      this.$store.commit('setError', {
        code: 0,
        msg: this.$t('misc.copied')
      })
    },
    onCopyFailed() {
      this.$store.commit('setError', {
        code: -1,
        msg: this.$t('misc.copy_failed')
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