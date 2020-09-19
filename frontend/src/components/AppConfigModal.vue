<template>
  <div class="modal active" @keydown.esc="close">
    <a class="modal-overlay" aria-label="Close" @click="close"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="close"></a>
        <div class="modal-title h5" v-t="'config.title'"></div>
      </div>
      <div class="modal-body p-0 bg-gray">
        <div class="content">
          <ul class="tab tab-block">
            <li class="tab-item"
                :class="{ active: active === 'options' }"
                @click="active = 'options'">
              <a v-t="'config.title'"></a>
            </li>
            <li class="tab-item"
                :class="{ active: active === 'about' }"
                @click="active = 'about'">
              <a v-t="'config.about.title'"></a>
            </li>
          </ul>
          <div class="p-2 m-2" v-if="active === 'options'">
            <div class="form-horizontal">
              <div class="form-group">
                <div class="col-3">
                  <label class="form-label" for="options-lang" v-t="'config.lang.label'"></label>
                </div>
                <div class="col-9">
                  <select id="options-lang"
                          class="form-select"
                          :value="lang"
                          @change="changeLanguage">
                    <option>en</option>
                    <option>zh-Hans</option>
                  </select>
                </div>
              </div>
            </div>
          </div>
          <div class="p-2 m-2" v-if="active === 'about'">
            // To be done...
          </div>
        </div>
      </div>
      <div class="modal-footer"></div>
    </div>
  </div>
</template>

<script>
export default {
  name: "AppConfigModal",
  computed: {
    lang() {
      return this.$root.$i18n.locale
    }
  },
  data: function() {
    return {
      active: 'options'
    }
  },
  methods: {
    close() {
      this.$emit('close')
    },
    changeLanguage(event) {
      this.$root.$i18n.locale = event.target.value
    }
  }
}
</script>

<style scoped>

</style>