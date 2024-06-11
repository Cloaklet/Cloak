<script setup lang="ts">
import { useGlobalStore } from '@/stores/global';
import { computed, onMounted, ref, nextTick } from 'vue';

const store = useGlobalStore();
const password = ref('');
const passwordInput = ref();
const selectedVault = computed(() => store.selectedVault)
const emit = defineEmits(['close', 'unlock-vault-request'])

const requestUnlockVault = () => {
  emit('unlock-vault-request', {
    vaultId: selectedVault.value?.id,
    password: password.value
  })
}

onMounted(() => nextTick(() => passwordInput.value.focus()))
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
      <div class="modal-body">
        <div class="content">
          <div class="form-group">
            <i18n tag="label" for="vault-password" class="form-label" path="panel.unlock.password.label">
              <template #vaultname>{{ selectedVault?.name }}</template>
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
                :class="{ loading: false }"
                @click="requestUnlockVault"
                v-t="'misc.unlock'"></button>
        <button class="btn ml-1" aria-label="Close" @click="$emit('close')" v-t="'misc.cancel'"></button>
      </div>
    </div>
  </div>
</template>


<style scoped>

</style>