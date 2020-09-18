<template>
  <div class="modal active" @keydown.esc="close">
    <a href="javascript:void(0)" class="modal-overlay" aria-label="Close" @click="close"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="close"></a>
        <div class="modal-title h5">{{ title }}</div>
      </div>
      <div class="modal-body p-0 bg-gray">
        <div class="content">
          <div class="menu p-0">
            <div class="menu-item mt-0" @click="clickOnParentPath" v-if="hasParent">
              <a href="javascript:void(0)">
                <div class="tile tile-centered">
                  <div class="tile-icon"><i class="ri-arrow-up-line"></i></div>
                  <div class="tile-content">
                    <div class="tile-title h6 text-normal">Go to parent directory</div>
                  </div>
                </div>
              </a>
            </div>
            <div class="menu-item mt-0" v-for="item in items" :key="item.name" @click="clickOnItem(item)">
              <a href="javascript:void(0)" :class="{ active: selected === item }">
                <div class="tile tile-centered">
                  <div class="tile-icon">
                    <i class="ri-file-shield-2-line" v-if="item.type === 'file'"></i>
                    <i class="ri-folder-5-line" v-else></i>
                  </div>
                  <div class="tile-content">
                    <div class="tile-title h6 text-normal">{{ item.name }}</div>
                  </div>
                </div>
              </a>
            </div>
            <div class="empty p-2" v-if="!items.length">
              <p class="empty-subtitle">Nothing here...</p>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer px-0">
        <input class="form-input float-left d-inline-block text-primary current-path"
               v-model="pwd"
               disabled>
        <button class="btn btn-primary"
                :disabled="!canFinishSelection"
                @click="selectItem">{{ okBtn }}</button>
        <button class="btn ml-1" aria-label="Close" @click="close">Cancel</button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import {mapMutations} from 'vuex'

const API = 'http://127.0.0.1:9763' // FIXME

export default {
  name: "FileSelectionModal",
  props: {
    mode: {
      type: String,
      required: true,
      validator: function(value) {
        return ['file', 'directory'].indexOf(value) !== -1
      }
    },
    title: {
      type: String,
      default: `Select an item`
    },
    okBtn: {
      type: String,
      default: "Select"
    }
  },
  data: function () {
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
    },
    canFinishSelection: function() {
      // Either we've selected a file in file mode, or we are inside a directory in directory mode
      return this.mode === 'directory' || (this.mode === 'file' && !!this.selected)
    },
  },
  methods: {
    ...mapMutations(['setError']),
    listSubPaths(pwd) {
      axios.post(`${API}/api/subpaths`, {
        pwd: pwd,
      }).then(resp => {
        if (resp.data.code !== 0) {
          // Pass error to root app via custom event
          return this.setError(resp.data)
        }
        this.pwd = resp.data.pwd;
        this.items = resp.data.items;
        this.sep = resp.data.sep;
      }).catch(err => {
        return this.setError({code: -1, msg: err.message}) // FIXME
      })
    },
    clickOnItem(item) {
      // Click on a directory enters it
      if (item.type === 'directory') {
        this.selected = null;
        return this.listSubPaths(`${this.pwd}${this.sep}${item.name}`)
      }
      // Click on a file selects it, but only if we're in file mode
      // Only allow selecting gocryptfs.conf
      if (this.mode === 'file' && item.name === 'gocryptfs.conf') {
        this.selected = item
      }
    },
    clickOnParentPath() {
      this.selected = null;
      this.listSubPaths(`${this.pwd}${this.sep}..`)
    },
    selectItem() {
      if (this.mode === 'file') {
        this.$emit('selected', `${this.pwd}${this.sep}${this.selected.name}`)
      } else if (this.mode === 'directory') {
        this.$emit('selected', `${this.pwd}`)
      }
      this.close()
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
}
</script>

<style scoped>
.modal-body {
  border: .05rem solid #dadee4;
  height: 450px;
}
.modal-body .content, .modal-body .content .menu {
  height: 100%;
}
.current-path {
  max-width: 75%;
}
</style>