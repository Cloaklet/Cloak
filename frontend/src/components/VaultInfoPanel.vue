<script setup lang="ts">
import VaultUnlockModal from "./VaultUnlockModal.vue";
import VaultOptionsModal from "./VaultOptionsModal.vue";
import { computed, ref } from 'vue';
import { useGlobalStore } from '@/stores/global';
import { useI18n } from "vue-i18n";

const showUnlock = ref(false);
const showVaultOptionsModal = ref(false);

const {t} = useI18n();
const store = useGlobalStore();

const selectedVault = computed(() => store.selectedVault)
const unlockVault = (payload: {vaultId: string, password: string}) => {
  store.unlockVault(payload).then(() => {
    showUnlock.value = false
  })
}
</script>
<template>
  <div class="column col-8 info-panel">
    <div class="empty" v-if="!selectedVault">
      <p class="empty-title h5" v-t="'panel.notselected.title'"></p>
      <p class="empty-subtitle" v-t="'panel.notselected.subtitle'"></p>
    </div>
    <div class="bg-gray vault-info" v-else>
      <div class="tile tile-centered">
        <div class="tile-icon text-center text-light bg-primary s-circle">
          <i class="ri-lock-unlock-fill ri-lg" v-if="selectedVault.state === 'unlocked'"></i>
          <i class="ri-lock-fill ri-lg" v-else></i>
        </div>
        <div class="tile-content p-relative">
          <div class="tile-title h5">{{ selectedVault.name }}</div>
          <small class="tile-subtitle text-gray" @click.stop="store.revealVault({vaultId: selectedVault.id})">
            {{ selectedVault.path }}
            <i class="ri-eye-fill vault-reveal-icon ml-1"/>
          </small>
          <span class="chip text-uppercase text-bold text-light p-absolute"
                :class="{ 'bg-primary': selectedVault.state === 'unlocked' }"
                v-t="selectedVault.state"></span>
        </div>
      </div><!--vault info end-->
      <div class="vault-operations text-center" v-if="selectedVault.state === 'unlocked'">
        <button class="btn btn-primary btn-lg"
                @click="store.revealMountPointForVault({vaultId: selectedVault.id})">
          <i class="ri-folder-open-fill"></i> {{ $t('panel.buttons.reveal') }}
        </button>
        <div class="text-center mt-2">
          <button class="btn btn-sm"
                  @click="store.lockVault({vaultId: selectedVault.id})">
            <i class="ri-key-2-fill"></i> {{ $t('panel.buttons.lock') }}
          </button>
        </div>
      </div><!--vault operation buttons end-->
      <div class="vault-operations text-center" v-else>
        <button class="btn btn-primary btn-lg"
                @click="showUnlock = true">
          <i class="ri-key-2-fill"></i> {{ $t('panel.buttons.unlock') }}
        </button>
        <div class="text-center mt-2">
          <button class="btn btn-sm text-dark"
                  @click="showVaultOptionsModal = true">
            <i class="ri-settings-3-fill"></i> {{ $t('panel.buttons.vault_options') }}
          </button>
        </div>
      </div>
    </div>
    <VaultUnlockModal v-if="showUnlock"
                      @close="showUnlock = false"
                      @unlock-vault-request="unlockVault"/>
    <VaultOptionsModal v-if="showVaultOptionsModal"
                       @close="showVaultOptionsModal = false" />
  </div>
</template>


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
  font-size: .7em;
  user-select: none;
}
.vault-info .s-circle {
  padding: .4rem;
  width: 40px;
  height: 40px;
}
.vault-info [class^=ri-] {
  vertical-align: baseline;
}
.vault-operations {
  margin-top: 3rem;
}
.vault-info .tile-title {
  user-select: none;
}
.vault-info .tile-subtitle {
  cursor: pointer;
}
.vault-info .tile-subtitle .vault-reveal-icon {
  display: none;
}
.vault-info .tile-subtitle:hover .vault-reveal-icon {
  display: inline-block;
}
</style>