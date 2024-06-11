<script setup lang="ts">
import { useGlobalStore } from '@/stores/global';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const {t, locale} = useI18n();

const store = useGlobalStore();
const version = computed(() => store.version);
const lang = computed(() => store.options.locale);
const loglevel = computed(() => store.options.loglevel);

const enum tab  {
  options = 'options',
  about = 'about',
}
const active = ref<tab>(tab.options);

const changeLanguage = (event: Event) => {
  store.setOptions({
    locale: (event.target as HTMLSelectElement).value
  })
  // FIXME frontend i18n
  locale.value = store.options.locale!
}
const changeLogLevel = (event: Event) => {
  store.setOptions({
    loglevel: (event.target as HTMLSelectElement).value
  })
}
</script>
<template>
  <div class="modal active" @keydown.esc="$emit('close')">
    <a class="modal-overlay" aria-label="Close" @click="$emit('close')"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="$emit('close')"></a>
        <div class="modal-title h5" v-t="'config.title'"></div>
      </div>
      <div class="modal-body p-2 bg-gray">
        <div class="content mx-2">
          <ul class="tab tab-block">
            <li class="tab-item"
                :class="{ active: active === tab.options }"
                @click="active = tab.options">
              <a><i class="ri-tools-fill"></i> {{ t('config.general.title') }}</a>
            </li>
            <li class="tab-item"
                :class="{ active: active === tab.about }"
                @click="active = tab.about">
              <a><i class="ri-information-fill"></i> {{ t('config.about.title') }}</a>
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
              <div class="form-group">
                <div class="col-3">
                  <label class="form-label" for="options-loglevel" v-t="'config.loglevel.label'"></label>
                </div>
                <div class="col-9">
                  <select id="options-loglevel"
                          class="form-select"
                          :value="loglevel"
                          @change="changeLogLevel">
                    <option>TRACE</option>
                    <option>DEBUG</option>
                    <option>INFO</option>
                    <option>WARN</option>
                    <option>ERROR</option>
                    <option>FATAL</option>
                    <option>PANIC</option>
                  </select>
                </div>
              </div>
            </div>
          </div>
          <div class="p-2 m-2" v-if="active === 'about'">
            <div class="tile">
              <div class="tile-content">
                <div class="tile-title text-bold">Cloak</div>
                <div class="tile-subtitle">{{ version.version || 'DEV' }} ({{ version.gitCommit || 'unknown' }})</div>
                <div class="tile-subtitle text-gray">Built@ {{ version.buildTime || 'unknown' }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer"></div>
    </div>
  </div>
</template>


<style scoped>

</style>