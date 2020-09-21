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
              <a><i class="ri-u-disk-fill"></i> {{ $t('vault.options.mounting.mounting') }}</a>
            </li>
            <li class="tab-item"
                :class="{ active: active === 'password' }"
                @click="active = 'password'">
              <a><i class="ri-key-2-fill"></i> {{ $t('vault.options.password') }}</a>
            </li>
          </ul>
          <div class="p-2 m-2" v-if="active === 'general'">
            <div class="form-horizontal pt-0">
              <div class="form-group">
                <div class="col-5">
                  <label class="form-label"
                         for="vault-options-autoreveal"
                         v-t="'vault.options.autoreveal.label'"></label>
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
            <div class="form-group">
              <label class="form-switch">
                <input type="checkbox"
                       :checked="selectedVault.readonly"
                       v-wait:disabled="'updating vault options'"
                       @change="updateVaultOptions({readonly: $event.target.checked})">
                <i class="form-icon"></i> {{ $t('vault.options.readonly.label') }}
              </label>
            </div>
            <div class="form-group">
              <label class="form-switch">
                <input type="checkbox"
                       v-model="useCustomMountpoint"
                       v-wait:disabled="'updating vault options'">
                <i class="form-icon"></i> {{ $t('vault.options.mounting.manual') }}
              </label>
            </div>
            <div class="form-group custom-mountpoint" v-if="useCustomMountpoint">
              <div class="input-group">
                <input class="form-input"
                       type="text"
                       v-model="selectedVault.mountpoint"
                       disabled>
                <button class="btn input-group-btn"
                        v-t="'select.file'"
                        @click="showMountpointSelectionModal = true"></button>
              </div>
            </div>
          </div>
          <div class="p-2 m-2" v-else-if="active === 'password'">
            <div class="text-center">
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
    <FileSelectionModal v-if="active === 'mount' && showMountpointSelectionModal"
                        :title="$t('vault.options.mounting.selection.title')"
                        :ok-btn="$t('misc.select')"
                        @close="mountpointSelectionClosed"
                        @selected="setVaultMountpoint"
                        mode="directory"/>
  </div>
</template>

<script>
import {mapGetters} from 'vuex'
import VaultPasswordChangingModal from "@/components/VaultPasswordChangingModal";
import FileSelectionModal from "@/components/FileSelectionModal";

export default {
  name: "VaultOptionsModal",
  components: {FileSelectionModal, VaultPasswordChangingModal},
  data: function() {
    return {
      active: 'general',
      showPasswordChangingModal: false,
      showMountpointSelectionModal: false,
      useCustomMountpoint: null
    }
  },
  watch: {
    useCustomMountpoint(newValue, oldValue) {
      // null => true/false means initialize (see mounted()), so ignore
      if (oldValue !== null) {
        // false => true means user enabling, so we need to show file selection modal
        if (newValue) {
          this.showMountpointSelectionModal = newValue
        } else { // true => false means user disabling, clearing if necessary
          if (this.selectedVault.mountpoint) {
            this.setVaultMountpoint('')
          }
        }
      }
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
        readonly: (typeof payload.readonly === 'undefined') ? this.selectedVault.readonly : payload.readonly,
        mountpoint: (typeof payload.mountpoint === 'undefined') ? this.selectedVault.mountpoint : payload.mountpoint
      }).finally(() => {
        this.$wait.end('updating vault options')
        this.showMountpointSelectionModal = false
      })
    },
    setVaultMountpoint(path) {
      console.log(path)
      this.updateVaultOptions({mountpoint: path})
    },
    mountpointSelectionClosed() {
      this.showMountpointSelectionModal = false
      this.useCustomMountpoint = !!this.selectedVault.mountpoint
    }
  },
  mounted() {
    this.useCustomMountpoint = !!this.selectedVault.mountpoint
  }
}
</script>

<style scoped>
.custom-mountpoint {
  padding-left: 2rem;
}
</style>