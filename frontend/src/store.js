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
        },
        minimalPasswordLength: () => {
            return 8
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
                autoreveal: payload.autoreveal,
                readonly: payload.readonly,
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
        },
        updateVault(state, payload) {
            for (let v of state.vaults) {
                if (v.id === payload.id) {
                    v.name = payload.path.split('/').pop()
                    v.path = payload.path
                    v.mountpoint = payload.mountpoint
                    v.autoreveal = payload.autoreveal
                    v.readonly = payload.readonly
                    v.state = payload.state
                    break
                }
            }
        }
    },
    actions: {
        // This method takes care of extra error handling when talking to backend APIs
        requestApi({commit}, payload) { // payload={method,api,data}
            let {method, api, data} = payload
            return axios.request({
                method: method,
                url: `${API}/api/${api}`,
                data: data
            }).then(resp => {
                if (resp.data.code !== 0) {
                    commit('setError', resp.data)
                }
                return resp.data
            }).catch(err => {
                commit('setError', {code: -1, msg: err.message}) // FIXME i18n
                return {}
            })
        },
        loadVaults ({commit, dispatch}) {
            return dispatch('requestApi', {
                method: 'get',
                api: 'vaults'
            }).then(({items}) => {
                let vaults = []
                for (const v of items || []) {
                    vaults.push({
                        id: v.id,
                        name: v.path.split('/').pop(),
                        path: v.path,
                        mountpoint: v.mountpoint,
                        autoreveal: v.autoreveal,
                        readonly: v.readonly,
                        state: v.state,
                        selected: false
                    })
                }
                commit('loadVaults', {vaults: vaults})
            })
        },
        removeVault({commit, dispatch}, {vaultId}) {
            return dispatch('requestApi', {
                method: 'delete',
                api: `vault/${vaultId}`}
            ).then(() => commit('removeVault', {vaultId: vaultId}))
        },
        addVault({commit, dispatch}, {path}) {
            return dispatch('requestApi', {
                method: 'post',
                api: 'vaults',
                data: {op: 'add', path: path}
            }).then(({item}) => item && commit('addVault', item))
        },
        createVault({commit, dispatch}, payload) { // payload={name,path,password}
            return dispatch('requestApi', {
                method: 'post',
                api: 'vaults',
                data: {
                    op: 'create',
                    ...payload
                }
            }).then(({item}) => item && commit('addVault', item))
        },
        revealMountPointForVault({dispatch}, {vaultId}) {
            return dispatch('requestApi', {
                method: 'post',
                api: `vault/${vaultId}`,
                data: {op: 'reveal'}
            })
        },
        lockVault({commit, dispatch}, {vaultId}) {
            return dispatch('requestApi', {
                method: 'post',
                api: `vault/${vaultId}`,
                data: {op: 'lock'}
            }).then(({state}) => {
                commit('setVaultState', {
                    vaultId: vaultId,
                    state: state || 'locked'
                })
            })
        },
        unlockVault({commit, dispatch}, {vaultId, password}) {
            return dispatch('requestApi', {
                method: 'post',
                api: `vault/${vaultId}`,
                data: {op: 'unlock', password: password}
            }).then(({state}) => commit('setVaultState', {
                vaultId: vaultId,
                state: state || 'locked'
            }))
        },
        updateVaultOptions({commit, dispatch}, payload) { // payload={vaultId,autoreveal,readonly,mountpoint}
            return dispatch('requestApi', {
                method: 'post',
                api: `vault/${payload.vaultId}/options`,
                data: {...payload}
            }).then(({item}) => item && commit('updateVault', item))
        },
        changeVaultPassword({dispatch}, payload) { // payload={vaultId,password,newpassword}
            return dispatch('requestApi', {
                method: 'post',
                api: `vault/${payload.vaultId}/password`,
                data: {...payload}
            })
        },
        revealVaultMasterkey({dispatch}, payload) { // payload={vaultId,password}
            return dispatch('requestApi', {
                method: 'post',
                api: `vault/${payload.vaultId}/masterkey`,
                data: {password: payload.password}
            }).then(({item}) => item)
        },
        loadAppConfig({dispatch}) {
            return dispatch('requestApi', {
                method: 'get',
                api: 'options'
            }).then(({item}) => {
                const version = item ? (item.version || {}) : {}
                return {
                    version: version.version || '',
                    gitCommit: version.gitCommit || '',
                    buildTime: version.buildTime || ''
                }
            })
        }
    }
})

