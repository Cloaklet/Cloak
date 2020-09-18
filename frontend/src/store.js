import Vuex from "vuex";
import axios from "axios";
import Vue from "vue";
Vue.use(Vuex)

const API = "http://127.0.0.1:9763"
export default new Vuex.Store({
    state: {
        vaults: [],
        error: {
            code: null,
            msg: '',
        }
    },
    getters: {
        selectedVault: state =>  {
            for (let v of state.vaults) {
                if (v.selected) {
                    return v
                }
            }
            return null
        },
        vaultsCount: state =>  {
            return state.vaults.length
        }
    },
    mutations: {
        loadVaults(state, payload) {
            Vue.set(state, 'vaults', payload.vaults)
        },
        setError(state, payload) {
            state.error.code = payload.code
            state.error.msg = payload.msg
        },
        closeAlert(state) {
            state.error.code = null
            state.error.msg = ''
        },
        selectVault(state, payload) {
            for (let v of state.vaults) {
                if (v.id === payload.vaultId) {
                    v.selected = true
                }
                if (v.selected && v.id !== payload.vaultId) {
                    v.selected = false
                }
            }
        },
        removeVault(state, payload) {
            for (let i = 0; i < state.vaults.length; i++) {
                if (state.vaults[i].id === payload.vaultId) {
                    state.vaults.splice(i, 1);
                    break
                }
            }
        },
        addVault(state, payload) {
            state.vaults.push({
                id: payload.id,
                name: payload.path.split('/').pop(),
                path: payload.path,
                mountpoint: payload.mountpoint,
                state: payload.state,
                selected: false,
            })
        },
        setVaultState(state, payload) {
            for (let v of state.vaults) {
                if (v.id === payload.vaultId) {
                    v.state = payload.state
                    break
                }
            }
        }
    },
    actions: {
        loadVaults ({commit}) {
            axios.get(`${API}/api/vaults`).then(resp => {
                if (resp.data.code !== 0) {
                    return commit('setError', resp.data)
                }
                let vaults = []
                for (const v of resp.data.items) {
                    vaults.push({
                        id: v.id,
                        name: v.path.split('/').pop(),
                        path: v.path,
                        mountpoint: v.mountpoint,
                        state: v.state,
                        selected: false
                    })
                }
                commit('loadVaults', {vaults: vaults})
            }).catch(err => {
                return commit('setError', {code: -1, msg: err.message}) // FIXME
            })
        },
        removeVault({commit}, payload) {
            axios.delete(`${API}/api/vault/${payload.vaultId}`).then(resp => {
                if (resp.data.code !== 0) {
                    return commit('setError', resp.data)
                }
                commit('removeVault', payload)
            }).catch(err => {
                return commit('setError', {code: -1, msg: err.message}) // FIXME
            })
        },
        addVault({commit}, payload) {
            axios.post(`${API}/api/vaults`, {
                op: 'add',
                path: payload.path
            }).then(resp => {
                if (resp.data.code !== 0) {
                    return commit('setError', resp.data)
                }
                commit('addVault', resp.data.item)
            }).catch(err => {
                return commit('setError', {code: -1, msg: err.message}) // FIXME
            })
        },
        createVault({commit}, payload) {
            axios.post(`${API}/api/vaults`, {
                op: 'create',
                path: payload.path,
                name: payload.name,
                password: payload.password
            }).then(resp => {
                if (resp.data.code !== 0) {
                    return commit('setError', resp.data)
                }
                commit('addVault', resp.data.item)
            }).catch(err => {
                return commit('setError', {code: -1, msg: err.message}) // FIXME
            })
        },
        revealMountPointForVault({commit}, payload) {
            axios.post(`${API}/api/vault/${payload.vaultId}`, {
                op: 'reveal'
            }).then(resp => {
                if (resp.data.code !== 0) {
                    return commit('setError', resp.data)
                }
            }).catch(err => {
                return commit('setError', {code: -1, msg: err.message}) // FIXME
            })
        },
        lockVault({commit}, payload) {
            axios.post(`${API}/api/vault/${payload.vaultId}`, {
                op: 'lock'
            }).then(resp => {
                if (resp.data.code !== 0) {
                    return commit('setError', resp.data)
                }
                commit('setVaultState', {
                    vaultId: payload.vaultId,
                    state: resp.data.state
                })
            }).catch(err => {
                return commit('setError', {code: -1, msg: err.message}) // FIXME
            })
        },
        unlockVault({commit}, payload) {
            axios.post(`${API}/api/vault/${payload.vaultId}`, {
                op: 'unlock',
                password: payload.password
            }).then(resp => {
                if (resp.data.code !== 0) {
                    return commit('setError', resp.data)
                }
                commit('setVaultState', {
                    vaultId: payload.vaultId,
                    state: resp.data.state
                })
            }).catch(err => {
                return commit('setError', {code: -1, msg: err.message}) // FIXME
            })
        }
    }
})
