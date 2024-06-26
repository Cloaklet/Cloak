<script setup lang="ts">
import { useGlobalStore } from '@/stores/global';
import FileSelectionModal from './FileSelectionModal.vue'
import PasswordStrengthMeter from './PasswordStrengthMeter.vue'
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const store = useGlobalStore();
const emit = defineEmits(['close', 'add-vault-request']);
const {t} = useI18n();

const mode = ref<null|'add'|'create'>(null);
const createVaultName = ref<string|null>(null);
const createVaultDir = ref<string|null>(null);
const addVaultFile = ref<string|null>(null);
const showDirSelection = ref(false);
const showFileSelection = ref(false);
const createVaultPassword = ref('');
const createVaultPasswordCheck = ref('');
const passwordStrengthHint = ref('');

const passwordMatch = computed(() => createVaultPassword.value === createVaultPasswordCheck.value);
const canCreate = computed(() => {
  return createVaultName.value &&
      createVaultDir.value &&
      createVaultPassword.value.length >= store.minimalPasswordLength &&
      passwordMatch.value
});

const setCreateVaultDir = (path: string) => {
  createVaultDir.value = path;
  showDirSelection.value = false;
}
const setAddVaultFile = (path: string) => {
  addVaultFile.value = path;
  showFileSelection.value = false;
}
const requestCreateVault = () => {
  store.createVault({
    name: createVaultName.value!,
    path: createVaultDir.value!,
    password: createVaultPassword.value,
  });
}
const requestAddVault = () => {
  emit('add-vault-request', {path: addVaultFile.value})
}
const showPasswordFeedback = ({warning}: {warning: string}) => {
  if (createVaultPassword.value.length < store.minimalPasswordLength) {
    passwordStrengthHint.value = t('misc.password.length_not_enough', {
      length: store.minimalPasswordLength
    })
    return
  }
  if (warning) {
    passwordStrengthHint.value = t(`zxcvbn.${warning}`) || warning
  } else {
    passwordStrengthHint.value = ''
  }
}

</script>
<template>
  <div class="modal active" @keydown.esc="$emit('close')">
    <a class="modal-overlay" aria-label="Close" @click="$emit('close')"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="$emit('close')"></a>
        <div class="modal-title h5" v-t="'list.add.title'"></div>
      </div>
      <div class="modal-body p-0 bg-gray">
        <div class="content m-2" :class="{ 'text-center': !mode }">
          <div class="p-2" v-if="!mode">
            <button class="btn btn-block my-2"
                    @click="mode = 'create'">
              <i class="ri-magic-line"></i> {{ $t('list.add.create') }}
            </button>
            <button class="btn btn-block my-2"
                    @click="mode = 'add'">
              <i class="ri-folder-open-line"></i> {{ $t('list.add.add') }}
            </button>
          </div>
          <div class="p-2" v-else>
            <div v-if="mode === 'create'">
              <FileSelectionModal v-if="showDirSelection"
                                  mode="directory"
                                  :title="$t('list.create.path.label')"
                                  :ok-btn="$t('misc.select')"
                                  @close="showDirSelection = false"
                                  @selected="setCreateVaultDir"/>
              <div class="form-group">
                <label for="create-vault-name"
                       class="form-label"
                       v-t="'list.create.name.label'"></label>
                <input type="text"
                       class="form-input"
                       id="create-vault-name"
                       :placeholder="$t('list.create.name.placeholder')"
                       v-model="createVaultName">
              </div>
              <div class="form-group">
                <label for="create-vault-directory"
                       class="form-label"
                       v-t="'list.create.path.label'"></label>
                <div class="input-group">
                  <input type="text"
                         class="form-input"
                         id="create-vault-directory"
                         :placeholder="$t('list.create.path.placeholder')"
                         v-model="createVaultDir"
                         disabled>
                  <button class="btn input-group-btn"
                          @click="showDirSelection = true"
                          v-t="'select.file'"></button>
                </div>
              </div>
              <div class="form-group" :class="{ 'has-error': passwordStrengthHint }">
                <label for="create-vault-password"
                       class="form-label"
                       v-t="'list.create.password.label'"></label>
                <PasswordStrengthMeter id="create-vault-password"
                                       v-model="createVaultPassword"
                                       :secure-length="store.minimalPasswordLength"
                                       :badge="false"
                                       :toggle="true"
                                       default-class="form-input"
                                       :label-show="$t('misc.show')"
                                       :label-hide="$t('misc.hide')"
                                       @feedback="showPasswordFeedback"/>
                <p class="form-input-hint"
                   v-if="passwordStrengthHint">{{ passwordStrengthHint }}</p>
              </div>
              <div class="form-group" :class="{ 'has-error': !passwordMatch }">
                <label for="create-vault-password-repeat"
                       class="form-label"
                       v-t="'list.create.password.repeat.label'"></label>
                <input type="password"
                       class="form-input"
                       id="create-vault-password-repeat"
                       v-model="createVaultPasswordCheck">
                <p class="form-input-hint"
                   v-if="!passwordMatch"
                   v-t="'list.create.password.repeat.notmatch'"></p>
              </div>
            </div>
            <div v-if="mode === 'add'">
              <FileSelectionModal v-if="showFileSelection"
                                  mode="file"
                                  :title="$t('list.add.select_file.title')"
                                  :ok-btn="$t('misc.open')"
                                  @close="showFileSelection = false"
                                  @selected="setAddVaultFile"/>
              <div class="form-group">
                <i18n tag="label" for="add-vault-file" class="form-label" path="list.add.select_file.label">
                  <template #filename>
                    <code>gocryptfs.conf</code>
                  </template>
                </i18n>
                <div class="input-group">
                  <input type="text"
                         class="form-input"
                         id="add-vault-file"
                         :placeholder="$t('list.add.vault_conf_file')"
                         v-model="addVaultFile"
                         disabled>
                  <button class="btn input-group-btn"
                          @click="showFileSelection = true"
                          v-t="'select.file'"></button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer px-0">
        <button class="btn float-left" v-if="mode" @click="mode = null" v-t="'misc.back'"></button>
        <button class="btn btn-primary float-right"
                v-if="mode === 'create'"
                :class="{ loading: false }"
                v-wait:click.start="'creating vault'"
                :disabled="!canCreate"
                @click="requestCreateVault"
                v-t="'misc.create'"></button>
        <button class="btn btn-primary float-right"
                v-if="mode === 'add'"
                :class="{ loading: false}"
                v-wait:click.start="'adding vault'"
                :disabled="!addVaultFile "
                @click="requestAddVault"
                v-t="'misc.add'"></button>
      </div>
    </div>
  </div>
</template>


<style scoped>
.form-input-hint {
  margin-bottom: 0;
}
* :deep(.Password__strength-meter) {
  margin: .4rem auto;
}
</style>
