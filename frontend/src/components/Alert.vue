<template>
  <div class="toast p-fixed d-inline"
       :class="[ isError ? 'toast-error': 'toast-success' ]"
       v-show="hasAlert">
    <button class="btn btn-clear float-right" v-if="isError" @click="closeAlert"></button>
    <p>{{ $store.state.error.msg }}
      <i18n tag="span" class="pl-1" v-if="isError" path="alert.errcode">
        <template #code>{{ errCode }})</template>
      </i18n>
    </p>
  </div>
</template>

<script>
import {mapMutations} from 'vuex'

export default {
  name: 'Alert',
  data: function () {
    return { timeoutId: null }
  },
  computed: {
    hasAlert() {
      return this.$store.state.error.msg.length > 0
    },
    isError() {
      return this.$store.state.error.code !== 0
    },
    errCode() {
      return this.$store.state.error.code
    }
  },
  methods: {
    ...mapMutations(['closeAlert'])
  },
  watch: {
    errCode(newValue) {
      // Use setTimeout to automatically close the alert if it is not an error.
      // If the code changes then we reset the timeout.
      if (this.timeoutId !== null) {
        clearTimeout(this.timeoutId);
        this.timeoutId = null
      }

      if (newValue === 0) {
        this.timeoutId = setTimeout(() => {
          this.closeAlert();
          this.timeoutId = null
        }, 2000)
      }
    }
  }
}
</script>

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
