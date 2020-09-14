Vue.prototype.$errorInfo = function(err) {
  if (err.response && err.response.data) {
    return [err.response.data.code, err.response.data.msg]
  }
  return [-1, err.message || 'Unknown error']
};

// Vault password prompt
Vue.component('vault-unlock-modal', {
  template: '#vault-unlock-modal-template',
  props: ['vault', 'unlocking'],
  delimiters: ['${', '}'],
  data: function () {
    return {
      password: "",
    }
  },
  methods: {
    close() {
      this.password = "";
      this.$emit('close')
    },
    requestUnlockVault() {
      this.$emit('unlock-vault', {id: this.vault.id, password: this.password})
    }
  },
  mounted() {
    this.$nextTick(() => this.$refs.passwordInput.focus())
  }
});

// File selection dialog
Vue.component('file-selection-modal', {
  template: '#file-selection-modal-template',
  delimiters: ['${', '}'],
  data: function() {
    return {
      pwd: "",
      items: [],
      selected: null,
      sep: "",  // path separator
    }
  },
  computed: {
    hasParent: function() {
      return this.pwd && this.pwd !== '/'
    }
  },
  methods: {
    listSubPaths(pwd) {
      axios.post(`/api/subpaths`, {
        pwd: pwd,
      }).then(resp => {
        if (resp.data.code !== 0) {
          // Pass error to root app via custom event
          return this.$emit('alert', resp.data.code, resp.data.msg)
        }
        this.pwd = resp.data.pwd;
        this.items = resp.data.items;
        this.sep = resp.data.sep;
      }).catch(err => {
        return this.$emit(...['alert'].concat(this.$root.$errorInfo(err)))
      })
    },
    selectItem(item) {
      if (item.type === 'directory') {
        this.selected = null;
        return this.listSubPaths(`${this.pwd}${this.sep}${item.name}`)
      }
      // Only allow selecting gocryptfs.conf
      if (item.name === 'gocryptfs.conf') {
        this.selected = item
      }
    },
    selectParentPath() {
      this.selected = null;
      this.listSubPaths(`${this.pwd}${this.sep}..`)
    },
    requestAddVault() {
      this.$emit('add-vault', `${this.pwd}`);
      this.$emit('close')
    },
    close() {
      this.$emit('close')
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.listSubPaths('$HOME')
    })
  }
});

// Alert message area
Vue.component('alert', {
  template: '#alert-template',
  delimiters: ['${', '}'],
  props: ['message', 'code'],
  watch: {
    code(newVaule) {
      // Use setTimeout to automatically close the alert if it is not an error.
      // If the code changes then we reset the timeout.
      if (this.timeoutId !== null) {
        clearTimeout(this.timeoutId);
        this.timeoutId = null
      }

      if (newVaule === 0) {
        this.timeoutId = setTimeout(() => {
          this.$emit('close-alert');
          this.timeoutId = null
        }, 2000)
      }
    },
  },
  data: function () {
    return {
      timeoutId: null,
    }
  },
  computed: {
    isError() {
      return this.code !== 0
    },
  },
});

new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    vaults: [],
    showUnlock: false,
    showFileSelection: false,
    unlocking: false,
    removing: false,
    errorCode: null,
    errorMessage: '',
  },
  computed: {
    selected: function () {
      for (let v of this.vaults) {
        if (v.selected) {
          return v
        }
      }
      return null
    }
  },
  methods: {
    alert(code, msg) {
      this.errorCode = code;
      this.errorMessage = msg
    },
    selectVault(vaultId) {
      for (let v of this.vaults) {
        if (v.id === vaultId) {
          v.selected = true
        }
        if (v.selected && v.id !== vaultId) {
          v.selected = false
        }
      }
    },
    addVault(vaultPath) {
      axios.post(`/api/vaults`, {
        op: 'add',
        path: vaultPath,
      }).then(resp => {
        if (resp.data.code !== 0) {
          return this.alert(resp.data.code, resp.data.msg)
        }
        let v = resp.data.item;
        this.vaults.push({
          id: v.id,
          name: v.path.split('/').pop(),
          path: v.path,
          mountpoint: v.mountpoint,
          state: v.state,
          selected: false,
        });
        this.alert(resp.data.code, resp.data.msg);
      }).catch(err => {
        return this.alert(...this.$errorInfo(err))
      })
    },
    removeVault(vaultId) {
      this.removing = true;
      axios.delete(`/api/vault/${vaultId}`).then(resp => {
        if (resp.data.code !== 0) {
          return this.alert(resp.data.code, resp.data.msg)
        }
        for (let i = 0; i < this.vaults.length; i++) {
          if (this.vaults[i].id === vaultId) {
            this.vaults.splice(i, 1);
            break
          }
        }
        this.alert(resp.data.code, resp.data.msg);
        this.removing = false;
      }).catch(err => {
        this.removing = false;
        return this.alert(...this.$errorInfo(err))
      })
    },
    unlockVault(info) { // info = {id, password}
      this.unlocking = true;
      axios.post(`/api/vault/${info.id}`, {
        op: 'unlock',
        password: info.password,
      }).then(resp => {
        this.unlocking = false;
        this.showUnlock = false;
        if (resp.data.code !== 0) {
          return this.alert(resp.data.code, resp.data.msg)
        }
        this.alert(resp.data.code, resp.data.msg);
        this.selected.state = resp.data.state
      }).catch(err => {
        this.unlocking = false;
        return this.alert(...this.$errorInfo(err))
      })
    },
    lockVault(vaultId) {
      axios.post(`/api/vault/${vaultId}`, {
        op: 'lock',
      }).then(resp => {
        if (resp.data.code !== 0) {
          return this.alert(resp.data.code, resp.data.msg)
        }
        this.alert(resp.data.code, resp.data.msg);
        this.selected.state = resp.data.state
      }).catch(err => {
        return this.alert(...this.$errorInfo(err))
      })
    },
    revealMountpoint(vaultId) {
      axios.post(`/api/vault/${vaultId}`, {
        op: 'reveal',
      }).then(resp => {
        if (resp.data.code !== 0) {
          return this.alert(resp.data.code, resp.data.msg)
        }
        this.alert(resp.data.code, resp.data.msg);
      }).catch(err => {
        return this.alert(...this.$errorInfo(err))
      })
    },
    closeAlert() {
      this.errorMessage = '';
      this.errorCode = null
    },
  },
  mounted () {
    axios.get('/api/vaults').then(resp => {
      if (resp.data.code !== 0) {
        return this.alert(resp.data.code, resp.data.msg)
      }
      let vaults = [];
      for (const v of resp.data.items) {
        vaults.push({
          id: v.id,
          name: v.path.split('/').pop(),
          path: v.path,
          mountpoint: v.mountpoint,
          state: v.state,
          selected: false,
        })
      }
      this.vaults = vaults;
    }).catch(err => {
      return this.alert(...this.$errorInfo(err))
    })
  }
});