import { defineStore } from "pinia";

type vault = {
    id: string,
    name: string,
    path: string,
    mountpoint: string,
    autoreveal: boolean,
    readonly: boolean,
    state: string,
    selected: boolean
}
type error = {
    code: number|null,
    msg: string
}
type appVersion = {
  version?: string,
  gitCommit?: string,
  buildTime?: string,
}
const enum logLevel  {
  TRACE = 'TRACE',
  DEBUG = 'DEBUG',
  INFO = 'INFO',
  WARN = 'WARN',
  ERROR = 'ERROR',
  FATAL = 'FATAL',
  PANIC = 'PANIC',
}
type appOptions = {
  locale?: string,
  loglevel?: logLevel,
}

const API = "http://127.0.0.1:9763"
const requestApi = ({method, api, data}: {
  method: string,
  api: string,
  data?: any, // TODO
}) => {
    return fetch(`${API}/api/${api}`, {
        method: method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: data ? JSON.stringify(data) : undefined,
    }).then(resp => {
      return resp.json().then(data => {
        if (data.code !== 0) {
          throw new Error(data.msg)
        }
        return data
      })
    })
}

export const useGlobalStore = defineStore('global', {
    state: () => ({
        vaults: [],
        error: {
            code: null,
            msg: '',
        },
        version: {},
        options: {}
    } as {
      vaults: vault[],
      error: error,
      version: appVersion,
      options: appOptions,
    }),
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
    actions: {
      selectVault({vaultId}:{vaultId: string}) {
        for (let v of this.vaults) {
          if (v.id === vaultId) {
            v.selected = true
          }
          if (v.selected && v.id !== vaultId) {
            v.selected = false
          }
        }
      },
        loadVaults () {
          requestApi({
            method: 'get',
            api: 'vaults',
          }).then(data => {
            this.vaults = data.items.map((v: vault) => ({
              id: v.id,
              name: v.path.split('/').pop(),
              path: v.path,
              mountpoint: v.mountpoint,
              autoreveal: v.autoreveal,
              readonly: v.readonly,
              state: v.state,
              selected: false
            }))
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        removeVault({vaultId}:{vaultId: string}) {
          requestApi({
            method: 'delete',
            api: `vault/${vaultId}`,
          }).then(() => {
            this.vaults = this.vaults.filter(v => v.id !== vaultId)
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        addVault({path}:{path: string}) {
          return requestApi({
            method: 'post',
            api: 'vaults',
            data: {op: 'add', path: path},
          }).then(data => {
            const v = data.item
            this.vaults.push({
              id: v.id,
              name: v.path.split('/').pop(),
              path: v.path,
              mountpoint: v.mountpoint,
              autoreveal: v.autoreveal,
              readonly: v.readonly,
              state: v.state,
              selected: false
            })
          })
        },
        createVault({name, path, password}: {
          name: string,
          path: string,
          password: string,
        }) { // payload={name,path,password}
          return requestApi({
            method: 'post',
            api: 'vaults',
            data: {op: 'create', name, path, password},
          }).then(data => {
            const v = data.item
            this.vaults.push({
              id: v.id,
              name: v.path.split('/').pop(),
              path: v.path,
              mountpoint: v.mountpoint,
              autoreveal: v.autoreveal,
              readonly: v.readonly,
              state: v.state,
              selected: false
            })
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        revealMountPointForVault({vaultId}:{vaultId: string}) {
          requestApi({
            method: 'post',
            api: `vault/${vaultId}`,
            data: {op: 'reveal'},
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        lockVault({vaultId}:{vaultId: string}) {
          requestApi({
            method: 'post',
            api: `vault/${vaultId}`,
            data: {op: 'lock'},
          }).then(data => {
            for (let v of this.vaults) {
              if (v.id === vaultId) {
                v.state = data.state
                break
              }
            }
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        unlockVault({vaultId, password}: {
          vaultId: string,
          password: string,
        }) {
          return requestApi({
            method: 'post',
            api: `vault/${vaultId}`,
            data: {op: 'unlock', password},
          }).then(data => {
            for (let v of this.vaults) {
              if (v.id === vaultId) {
                v.state = data.state
                break
              }
            }
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        updateVaultOptions(payload: {
          vaultId: string,
          autoreveal?: boolean,
          readonly?: boolean,
          mountpoint?: string,
        }) { // payload={vaultId,autoreveal,readonly,mountpoint}
          if (!payload.autoreveal && !payload.readonly && !payload.mountpoint) {
            return
          }
          return requestApi({
            method: 'post',
            api: `vault/${payload.vaultId}/options`,
            data: {...payload},
          }).then(data => {
            const vault: vault = data.item
            for (let v of this.vaults) {
              if (v.id === payload.vaultId) {
                v.name = vault.mountpoint.split('/').pop()!
                v.path = vault.path
                v.autoreveal = vault.autoreveal
                v.readonly = vault.readonly
                v.mountpoint = vault.mountpoint
                v.state = vault.state
                break
              }
            }
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        changeVaultPassword(payload: {
          vaultId: string,
          password?: string,
          masterkey?: string,
          newpassword: string,
        }) { // payload={vaultId,password/masterkey,newpassword}
          return requestApi({
            method: 'post',
            api: `vault/${payload.vaultId}/password`,
            data: {...payload},
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        revealVaultMasterkey(payload: {
          vaultId: string,
          password: string,
        }) { // payload={vaultId,password}
          return requestApi({
            method: 'post',
            api: `vault/${payload.vaultId}/masterkey`,
            data: {password: payload.password},
          }).then(data => data.item).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        loadAppConfig() {
          return requestApi({
            method: 'get',
            api: 'options',
          }).then(({item}) => {
            this.version = item.version
            this.options = item.options
            return item.options
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        listSubPaths({path}: {path: string}) {
          return requestApi({
            method: 'post',
            api: 'subpaths',
            data: {pwd: path},
          }).then(data => {
            return data.items
          }).catch(e => {
            this.error = {code: -1, msg: e.message}
          })
        },
        setOptions(payload: {
          locale?: string,
          loglevel?: string,
        }) {
          if (!payload.locale && !payload.loglevel) {
            return
          }
            const data: any = {}
            if (payload.locale) {
                data.locale = payload.locale
            }
            if (payload.loglevel) {
                data.loglevel = payload.loglevel
            }
            requestApi({
                method: 'post',
                api: 'options',
                data: data
            }).then(() => {
                this.options = {...data}
            })
        },
        closeAlert() {
            this.error = {code: null, msg: ''}
        }
    },
})
