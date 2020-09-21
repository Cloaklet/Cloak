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
      <div class="modal-body p-2 bg-gray">
        <div class="content mx-2">
          <ul class="tab tab-block">
            <li class="tab-item"
                :class="{ active: active === 'general' }"
                @click="active = 'general'">
              <a><i class="ri-tools-fill"></i> {{ $t('config.general.title') }}</a>
            </li>
            <li class="tab-item"
                :class="{ active: active === 'mount' }"
                @click="active = 'mount'">
              <a><i class="ri-u-disk-fill"></i> {{ $t('vault.options.mounting') }}</a>
            </li>
            <li class="tab-item"
                :class="{ active: active === 'password' }"
                @click="active = 'password'">
              <a><i class="ri-key-2-fill"></i> {{ $t('vault.options.password') }}</a>
            </li>
          </ul>
          <div class="p-2 m-2" v-if="active === 'general'">
            <div class="form-horizontal py-2">
              <div class="form-group">
                <div class="col-5">
                  <label class="form-label" for="vault-options-autoreveal" v-t="'vault.options.autoreveal.label'"></label>
                </div>
                <div class="col-7">
                  <select id="vault-options-autoreveal"
                          class="form-select"
                          :value="selectedVault.autoreveal"
                          v-wait:disabled="'updating vault options'"
                          @change="updateVaultOptions({autoreveal: $event.target.value === 'true'})">
                    <option value="false" v-t="'vault.options.autoreveal.do_nothing'"></option>
                    <option value="true" v-t="'vault.options.autoreveal.reveal_drive'"></option>
                  </select>
                </div>
              </div>
            </div>
          </div>
          <div class="p-2 m-2" v-else-if="active === 'mount'">
            <div class="form-group py-2">
              <label class="form-checkbox">
                <input type="checkbox"
                       :checked="selectedVault.readonly"
                       v-wait:disabled="'updating vault options'"
                       @change="updateVaultOptions({readonly: $event.target.checked})">
                <i class="form-icon"></i> {{ $t('vault.options.readonly.label') }}
              </label>
            </div>
          </div>
          <div class="p-2 m-2" v-else-if="active === 'password'">
            <div class="text-center py-2">
              <button class="btn btn-block" @click="showPasswordChangingModal = true">
                <i class="ri-key-2-fill"></i> {{ $t('vault.options.buttons.change_password') }}
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer"></div>
    </div>
    <VaultPasswordChangingModal v-if="showPasswordChangingModal"
                                @close="showPasswordChangingModal = false"/>
  </div>
</template>

<script>
import {mapGetters} from 'vuex'
import VaultPasswordChangingModal from "@/components/VaultPasswordChangingModal";

export default {
  name: "VaultOptionsModal",
  components: {VaultPasswordChangingModal},
  data: function() {
    return {
      active: 'general',
      showPasswordChangingModal: false
    }
  },
  computed: {
    ...mapGetters(['selectedVault'])
  },
  methods: {
    close() {
      this.$emit('close')
    },
    updateVaultOptions(payload) {
      this.$wait.start('updating vault options')
      this.$store.dispatch('updateVaultOptions', {
        vaultId: this.selectedVault.id,
        autoreveal: (typeof payload.autoreveal === 'undefined') ? this.selectedVault.autoreveal : payload.autoreveal,
        readonly: (typeof payload.readonly === 'undefined') ? this.selectedVault.readonly : payload.readonly
      }).finally(() => {
        this.$wait.end('updating vault options')
      })
    }
  }
}
</script>

<style scoped>

</style>