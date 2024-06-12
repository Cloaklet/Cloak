<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useGlobalStore } from '@/stores/global';
import { useI18n } from 'vue-i18n';

const {te, t} = useI18n();
const timeoutId = ref<number|null>(null);
const store = useGlobalStore();
const hasAlert = computed(() => store.error.msg.length > 0);
const isError = computed(() => store.error.code !== 0);
const errCode = computed(() => store.error.code);
const translatedErrMsg = computed(() => {
  const key = `errors.api_${errCode.value}`;
  if (errCode.value! >= 0 && te(key)) {
    return t(key);
  }
  return store.error.msg;
});
const closeAlert = () => {
  store.error = {
    code: 0,
    msg: '',
  };
}
watch(errCode, (newValue, oldValue) => {
  if (oldValue !== null) {
    clearTimeout(timeoutId.value!);
    timeoutId.value = null;
  }
  if (newValue === 0) {
    timeoutId.value = window.setTimeout(() => {
      store.closeAlert();
      timeoutId.value = null;
    }, 2000);
  }
})
</script>

<template>
  <div class="toast p-fixed d-inline"
       :class="[ isError ? 'toast-error': 'toast-success' ]"
       v-show="hasAlert">
    <button class="btn btn-clear float-right" v-if="isError" @click="closeAlert"></button>
    <p>{{ translatedErrMsg }}
      <i18n tag="span" class="pl-1" v-if="isError" path="alert.errcode">
        <template #code>{{ errCode }}</template>
      </i18n>
    </p>
  </div>
</template>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
div {
  top: 1rem;
  transform: translate(-50%, 0);
  left: 50%;
  width: 50%;
  box-shadow: 0 .25rem 0.5rem rgba(48, 55, 66, .3);
  z-index: 999;
}
</style>
