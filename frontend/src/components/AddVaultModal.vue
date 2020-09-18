<template>
  <div class="modal active" @keydown.esc="close">
    <a class="modal-overlay" aria-label="Close" @click="close"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="close"></a>
        <div class="modal-title h5">Add Vault</div>
      </div>
      <div class="modal-body p-0 bg-gray">
        <div class="content m-2" :class="{ 'text-center': !mode }">
          <div class="p-2" v-if="!mode">
            <button class="btn btn-block my-2"
                    @click="mode = 'create'"><i class="ri-magic-line"></i> Create New Vault</button>
            <button class="btn btn-block my-2"
                    @click="mode = 'add'"><i class="ri-folder-open-line"></i> Add Existing Vault</button>
          </div>
          <div class="p-2" v-else>
            <div v-if="mode === 'create'">
              <FileSelectionModal v-if="showDirSelection"
                                  mode="directory"
                                  title="Select where to store this vault"
                                  ok-btn="Select"
                                  @close="showDirSelection = false"
                                  @selected="setCreateVaultDir"/>
              <div class="form-group">
                <label for="create-vault-name" class="form-label">Choose a name for the new vault</label>
                <input type="text"
                       class="form-input"
                       id="create-vault-name"
                       placeholder="Vault Name"
                       v-model="createVaultName">
              </div>
              <div class="form-group">
                <label for="create-vault-directory" class="form-label">Choose where to store this vault</label>
                <div class="input-group">
                  <input type="text"
                         class="form-input"
                         id="create-vault-directory"
                         placeholder="Vault Location"
                         v-model="createVaultDir"
                         disabled>
                  <button class="btn input-group-btn" @click="showDirSelection = true">Choose...</button>
                </div>
              </div>
              <div class="form-group">
                <label for="create-vault-password" class="form-label">Choose a password for the new vault</label>
                <input type="password"
                       class="form-input"
                       id="create-vault-password"
                       v-model="createVaultPassword">
              </div>
              <div class="form-group" :class="{ 'has-error': !passwordMatch }">
                <label for="create-vault-password-repeat"
                       class="form-label">Type the vault password again</label>
                <input type="password"
                       class="form-input"
                       id="create-vault-password-repeat"
                       v-model="createVaultPasswordCheck">
                <p class="form-input-hint" v-if="!passwordMatch">Password does not match</p>
              </div>
            </div>
            <div v-if="mode === 'add'">
              <FileSelectionModal v-if="showFileSelection"
                                  mode="file"
                                  title="Select the gocryptfs.conf file inside target vault"
                                  ok-btn="Open"
                                  @close="showFileSelection = false"
                                  @selected="setAddVaultFile"/>
              <div class="form-group">
                <label for="add-vault-file" class="form-label">
                  Choose the <code>gocryptfs.conf</code> file of your existing vault
                </label>
                <div class="input-group">
                  <input type="text"
                         class="form-input"
                         id="add-vault-file"
                         placeholder="Vault Config File"
                         v-model="addVaultFile"
                         disabled>
                  <button class="btn input-group-btn" @click="showFileSelection = true">Choose...</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer px-0">
        <button class="btn float-left" v-if="mode" @click="mode = null">Back</button>
        <button class="btn btn-primary float-right"
                v-if="mode === 'create'"
                :class="{ loading: $wait.is('creating vault') }"
                v-wait:disabled="'creating vault'"
                v-wait:click.start="'creating vault'"
                :disabled="!canCreate"
                @click="requestCreateVault">Create</button>
        <button class="btn btn-primary float-right"
                v-if="mode === 'add'"
                :class="{ loading: $wait.is('adding vault') }"
                v-wait:disabled="'adding vault'"
                v-wait:click.start="'adding vault'"
                :disabled="!addVaultFile"
                @click="requestAddVault">Add</button>
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
    },
    setAddVaultFile(path) {
      this.addVaultFile = path
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