<template>
  <div class="modal active" @keydown.esc="close">
    <a class="modal-overlay" aria-label="Close" @click="close"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="close"></a>
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
              <div class="form-group">
                <label for="create-vault-password"
                       class="form-label"
                       v-t="'list.create.password.label'"></label>
                <input type="password"
                       class="form-input"
                       id="create-vault-password"
                       v-model="createVaultPassword">
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
                :class="{ loading: $wait.is('creating vault') }"
                v-wait:click.start="'creating vault'"
                :disabled="!canCreate || $wait.is('creating vault')"
                @click="requestCreateVault"
                v-t="'misc.create'"></button>
        <button class="btn btn-primary float-right"
                v-if="mode === 'add'"
                :class="{ loading: $wait.is('adding vault') }"
                v-wait:click.start="'adding vault'"
                :disabled="!addVaultFile || $wait.is('adding vault')"
                @click="requestAddVault"
                v-t="'misc.add'"></button>
      </div>
    </div>
  </div>
</template>

<script>
import FileSelectionModal from './FileSelectionModal'

export default {
  name: "AddVaultModal",
  components: {
    FileSelectionModal
  },
  data: function() {
    return {
      mode: null,  // add / create / null
      createVaultName: null,
      createVaultDir: null,
      addVaultFile: null,
      showDirSelection: false,
      showFileSelection: false,
      createVaultPassword: '',
      createVaultPasswordCheck: '',
    }
  },
  computed: {
    passwordMatch() {
      return this.createVaultPassword === this.createVaultPasswordCheck
    },
    canCreate() {
      return this.createVaultName && this.createVaultDir && this.createVaultPassword && this.passwordMatch
    },
  },
  methods: {
    setCreateVaultDir(path) {
      this.createVaultDir = path
      this.showDirSelection = false
    },
    setAddVaultFile(path) {
      this.addVaultFile = path
      this.showFileSelection = false
    },
    close() {
      this.$emit('close')
    },
    requestCreateVault() {
      this.$emit('create-vault-request', {
        path: this.createVaultDir,
        name: this.createVaultName,
        password: this.createVaultPassword
      })
    },
    requestAddVault() {
      this.$emit('add-vault-request', {path: this.addVaultFile})
    },
  },
}
</script>

<style scoped>

</style>