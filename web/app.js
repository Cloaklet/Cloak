Vue.component("vault-unlock-modal", {
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
    }
  }
});

new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    vaults: [],
    showUnlock: false,
    unlocking: false,
  },
  computed: {
    selected: function () {
      for (const v of this.vaults) {
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
    unlockVault(info) {
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