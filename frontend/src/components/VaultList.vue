<script setup lang="ts">
import AddVaultModal from './AddVaultModal.vue'
import {useGlobalStore} from '@/stores/global'
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const {locale} = useI18n();
const store = useGlobalStore();
const showAddVaultModal = ref(false);

const selectedVault = computed(() => store.selectedVault);
const vaultsCount = computed(() => store.vaults.length);

const addVault = (payload: {path: string}) => {
  store.addVault(payload).then(() => {
    showAddVaultModal.value = false
  })
}
const createVault = (payload: {name: string, path: string, password: string}) => {
  store.createVault(payload).then(() => {
    showAddVaultModal.value = false
  })
}

onMounted(() => {
  store.loadAppConfig().then((options) => {
    locale.value = options.locale
  })
  store.loadVaults()
})

</script>

<template>
  <div class="column col-4 p-relative">
    <div class="menu p-0 vault-list">
      <div class="empty" v-if="!vaultsCount">
        <p class="empty-title h5" v-t="'list.novault.title'"></p>
        <p class="empty-subtitle" v-t="'list.novault.subtitle'"></p>
      </div>
      <div class="menu-item mt-0" v-for="vault in store.vaults" :key="vault.id">
        <a :class="{ active: vault.selected }" @click="store.selectVault({vaultId: vault.id})">
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
              :data-tooltip="$t('list.buttons.add')"
              @click="showAddVaultModal = true">➕</button>
      <button class="btn btn-lg bg-gray h6 text-normal tooltip"
              :data-tooltip="$t('list.buttons.remove')"
              :disabled="!store.selectedVault"
              :class="{ loading: false }"
              @click="store.removeVault({vaultId: selectedVault!.id})">➖</button>
    </div>
    <AddVaultModal v-if="showAddVaultModal"
                   @close="showAddVaultModal = false"
                   @add-vault-request="addVault"
                   @create-vault-request="createVault"/>
  </div>
</template>

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