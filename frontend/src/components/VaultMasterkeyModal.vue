<script setup lang="ts">
import { useGlobalStore } from '@/stores/global';
import { computed, nextTick, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const {t} = useI18n();
const store = useGlobalStore();

const passwordInput = ref();
const password = ref('');
const masterkey = ref('');

const selectedVault = computed(() => store.selectedVault);

const requestRevealMasterkey = () => {
  if(!selectedVault.value) {
    return
  }
  store.revealVaultMasterkey({
    vaultId: selectedVault.value.id,
    password: password.value,
  }).then((v: string) => {
    masterkey.value = v
  })
}
const onCopySucceeded = () => {
  store.error = {
    code: 0,
    msg: t('misc.copied'),
  }
}
const onCopyFailed = () => {
  store.error = {
    code: -1,
    msg: t('misc.copy_failed'),
  }
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
        <div class="modal-title h5" v-t="'vault.options.masterkey.title'"></div>
      </div>
      <div class="modal-body">
        <div class="content">
          <div class="form-group" v-if="!masterkey">
            <i18n-t tag="label" for="vault-password" class="form-label" keypath="panel.unlock.password.label">
              <template #vaultname>{{ selectedVault?.name }}</template>
            </i18n-t>
            <input type="password"
                   class="form-input"
                   :disabled="false"
                   id="vault-password"
                   ref="passwordInput"
                   v-model="password"
                   @keydown.enter="requestRevealMasterkey">
          </div>
          <div v-else>
            <i18n-t tag="p" class="form-group" keypath="vault.options.masterkey.description">
              <template #vaultname>{{ selectedVault?.name }}</template>
            </i18n-t>
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
                :disabled="!password.length "
                :class="{ loading: false }"
                @click="requestRevealMasterkey"
                v-t="'vault.options.masterkey.view'"></button>
        <button class="btn ml-1"
                v-if="!masterkey"
                aria-label="Close"
                @click="$emit('close')"
                v-t="'misc.cancel'"></button>
        <button class="btn ml-1" v-else @click="$emit('close')" v-t="'misc.done'"></button>
      </div>
    </div>
  </div>
</template>


<style scoped>

</style>