<script setup lang="ts">
import VaultPasswordChangingModal from "./VaultPasswordChangingModal.vue";
import FileSelectionModal from "./FileSelectionModal.vue";
import VaultMasterkeyModal from "./VaultMasterkeyModal.vue";
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { useGlobalStore } from '@/stores/global';
import { useI18n } from "vue-i18n";

const enum tab  {
  general = 'general',
  mount = 'mount',
  password = 'password',
}

const {t} = useI18n();
const store = useGlobalStore();

const active = ref<tab>(tab.general);
const showPasswordChangingModal = ref(false);
const showMasterkeyModal = ref(false);
const showPasswordRecoveryModal = ref(false);
const showMountpointSelectionModal = ref(false);
const useCustomMountpoint = ref<boolean | null>(null);

const selectedVault = computed(() => store.selectedVault);

const updateVaultOptions = (payload: {autoreveal?: boolean, readonly?: boolean, mountpoint?: string}) => {
  if (!selectedVault.value ||typeof selectedVault.value === 'undefined'){return}
  store.updateVaultOptions({
    vaultId: selectedVault.value.id,
    autoreveal: payload.autoreveal,
    readonly: payload.readonly ,
    mountpoint: payload.mountpoint,
  })?.then(() => {
    showMountpointSelectionModal.value = false
  })
}

watch(useCustomMountpoint, (newValue, oldValue) => {
  if (!selectedVault.value) {return}
  if (oldValue !== null) {
    if (newValue) {
      showMountpointSelectionModal.value = newValue
    } else {
      if (selectedVault.value.mountpoint) {
        setVaultMountpoint('')
      }
    }
  }
})
const setVaultMountpoint = (path: string) => {
  updateVaultOptions({mountpoint: path})
}
const mountpointSelectionClosed = () => {
  showMountpointSelectionModal.value = false
  useCustomMountpoint.value = selectedVault.value && !!selectedVault.value.mountpoint
}
onMounted(() => {
  nextTick(() => {
    useCustomMountpoint.value = selectedVault.value && !!selectedVault.value.mountpoint
  })
})

</script>
<template>
  <div class="modal active" @keydown.esc="$emit('close')">
    <a class="modal-overlay" aria-label="Close" @click="$emit('close')"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="$emit('close')"></a>
        <div class="modal-title h5">{{ selectedVault?.name }}</div>
      </div>
      <div class="modal-body p-2 bg-gray">
        <div class="content mx-2">
          <ul class="tab tab-block">
            <li class="tab-item"
                :class="{ active: active === 'general' }"
                @click="active = tab.general">
              <a><i class="ri-tools-fill"></i> {{ $t('config.general.title') }}</a>
            </li>
            <li class="tab-item"
                :class="{ active: active === 'mount' }"
                @click="active = tab.mount">
              <a><i class="ri-u-disk-fill"></i> {{ $t('vault.options.mounting.mounting') }}</a>
            </li>
            <li class="tab-item"
                :class="{ active: active === 'password' }"
                @click="active = tab.password">
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
                          :value="selectedVault?.autoreveal"
                          v-wait:disabled="'updating vault options'"
                          @change="updateVaultOptions({autoreveal: ($event.target as HTMLInputElement).value === 'true'})">
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
                       :checked="selectedVault?.readonly"
                       v-wait:disabled="'updating vault options'"
                       @change="updateVaultOptions({readonly: ($event.target as HTMLInputElement).checked})">
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
                       v-model="selectedVault!.mountpoint"
                       disabled>
                <button class="btn input-group-btn"
                        v-t="'select.file'"
                        @click="showMountpointSelectionModal = true"></button>
              </div>
            </div>
          </div>
          <div class="p-2 m-2" v-else-if="active === 'password'">
            <div class="form-group">
              <button class="btn btn-block" @click="showPasswordChangingModal = true">
                <i class="ri-key-2-fill"></i> {{ $t('vault.options.buttons.change_password') }}
              </button>
            </div>
            <div class="form-group">
              <button class="btn btn-block" @click="showMasterkeyModal = true">
                <i class="ri-eye-fill"></i> {{ $t('vault.options.buttons.reveal_masterkey') }}
              </button>
            </div>
            <div class="form-group">
              <button class="btn btn-block" @click="showPasswordRecoveryModal = true">
                <i class="ri-device-recover-fill"></i> {{ $t('vault.options.buttons.recover_password') }}
              </button>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer"></div>
    </div>
    <VaultPasswordChangingModal using="password"
                                v-if="showPasswordChangingModal"
                                @close="showPasswordChangingModal = false"/>
    <VaultPasswordChangingModal using="masterkey"
                                v-if="showPasswordRecoveryModal"
                                @close="showPasswordRecoveryModal = false"/>
    <FileSelectionModal v-if="active === 'mount' && showMountpointSelectionModal"
                        :title="$t('vault.options.mounting.selection.title')"
                        :ok-btn="$t('misc.select')"
                        @close="mountpointSelectionClosed"
                        @selected="setVaultMountpoint"
                        mode="directory"/>
    <VaultMasterkeyModal v-if="showMasterkeyModal" @close="showMasterkeyModal = false"/>
  </div>
</template>


<style scoped>
* {
  user-select: none;
}
.custom-mountpoint {
  padding-left: 2rem;
}
</style>