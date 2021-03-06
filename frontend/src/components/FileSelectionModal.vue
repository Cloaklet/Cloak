<template>
  <div class="modal active" @keydown.esc="close">
    <a class="modal-overlay" aria-label="Close" @click="close"></a>
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
              <a>
                <div class="tile tile-centered">
                  <div class="tile-icon"><i class="ri-arrow-up-line"></i></div>
                  <div class="tile-content">
                    <div class="tile-title h6 text-normal" v-t="'select.gotoparent'"></div>
                  </div>
                </div>
              </a>
            </div>
            <div class="menu-item mt-0" v-for="item in items" :key="item.name" @click="clickOnItem(item)">
              <a :class="{ active: selected === item }">
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
              <p class="empty-subtitle" v-t="'select.empty'"></p>
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
        <button class="btn ml-1" aria-label="Close" @click="close" v-t="'misc.cancel'"></button>
      </div>
    </div>
  </div>
</template>

<script>
import {mapMutations} from 'vuex'

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
      default: function() {
        return this.$t('select.default_title')
      }
    },
    okBtn: {
      type: String,
      default: function() {
        return this.$t('misc.select')
      }
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
      this.$store.dispatch('listSubPaths', {path: pwd}).then(data => {
        this.pwd = data.pwd
        this.sep = data.sep
        this.items = data.items
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
.modal-body .content {
  min-height: 100%;
  display: grid; /* any better idea on how to stretch file list height? */
}
.modal-body .content .menu, .menu .empty {
  height: 100%;
}
.current-path {
  max-width: 75%;
}
</style>