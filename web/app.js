Vue.component('vault-unlock-modal', {
  template: '#vault-unlock-modal-template',
  props: ['vault'],
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
  methods: {
    listSubPaths(pwd) {
      axios.post(`/api/subpaths`, {
        pwd: pwd,
      }).then(resp => {
        if (resp.data.code !== 0) {
          return alert(JSON.stringify(resp)) // FIXME
        }
        this.pwd = resp.data.pwd;
        this.items = resp.data.items;
        this.sep = resp.data.sep;
      }).catch(err => {
        console.error(err)
        // FIXME
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

new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    vaults: [],
    showUnlock: false,
    showFileSelection: false,
    unlocking: false,
    removing: false,
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
          return alert(JSON.stringify(resp)) // FIXME
        }
        // TODO
        let v = resp.data.item;
        this.vaults.push({
          id: v.id,
          name: v.path.split('/').pop(),
          path: v.path,
          mountpoint: v.mountpoint,
          state: v.state,
          selected: false,
        })
      })
    },
    removeVault(vaultId) {
      this.removing = true;
      axios.delete(`/api/vault/${vaultId}`).then(resp => {
        this.removing = false;
        if (resp.data.code !== 0) {
          return alert(JSON.stringify(resp)) // FIXME
        }
        for (let i = 0; i < this.vaults.length; i++) {
          if (this.vaults[i].id === vaultId) {
            this.vaults.splice(i, 1);
            break
          }
        }
      }).catch((err) => {
        this.removing = false
      })
    },
    unlockVault(info) { // info = {id, password}
      this.unlocking = true;
      axios.post(`http://127.0.0.1:9763/api/vault/${info.id}`, {
        op: 'unlock',
        password: info.password,
      }).then(resp => {
        this.unlocking = false;
        this.showUnlock = false;
        if (resp.data.code !== 0) {
          return alert(JSON.stringify(resp)) // FIXME
        }
        this.selected.state = resp.data.state
      }).catch((err) => {
        this.unlocking = false
      })
    },
    lockVault(vaultId) {
      axios.post(`http://127.0.0.1:9763/api/vault/${vaultId}`, {
        op: 'lock',
      }).then(resp => {
        if (resp.data.code !== 0) {
          return alert(JSON.stringify(resp)) // FIXME
        }
        this.selected.state = resp.data.state
      })
    },
    revealMountpoint(vaultId) {
      // FIXME
      console.log(`TBD: revealing mountpoint for ${vaultId}`)
    }
  },
  mounted () {
    axios.get('http://127.0.0.1:9763/api/vaults').then(resp => {
      if (resp.data.code !== 0) {
        return alert(JSON.stringify(resp)) // FIXME
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
    })
  }
});