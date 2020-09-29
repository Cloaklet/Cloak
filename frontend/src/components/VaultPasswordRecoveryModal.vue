<template>
  <div class="modal active" @keydown.esc="close">
    <a class="modal-overlay" aria-label="Close" @click="close"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="close"></a>
        <div class="modal-title h5" v-t="'vault.options.recover_password.title'"></div>
      </div>
      <div class="modal-body">
        <div class="content">
          <div class="form-group" :class="{ 'has-error': !masterkeyValid }">
            <i18n tag="label"
                  class="form-label"
                  for="vault-recoverpw-masterkey"
                  path="vault.options.recover_password.label.masterkey">
              <template #vaultname>{{ selectedVault.name }}</template>
            </i18n>
            <input class="form-input"
                   type="text"
                   id="vault-recoverpw-masterkey"
                   ref="masterkeyInput"
                   v-model="password">
          </div>
          <div class="form-group" :class="{ 'has-error': passwordStrengthHint }">
            <label class="form-label"
                   for="vault-recoverpw-newpassword"
                   v-t="'vault.options.change_password.label.newpassword'"></label>
            <PasswordStrengthMeter id="vault-recoverpw-newpassword"
                                   v-model="newPassword"
                                   :secure-length="$store.getters.minimalPasswordLength"
                                   :badge="false"
                                   :toggle="true"
                                   default-class="form-input"
                                   :label-show="$t('misc.show')"
                                   :label-hide="$t('misc.hide')"
                                   @feedback="showPasswordFeedback"/>
            <p class="form-input-hint"
               v-if="passwordStrengthHint">{{ passwordStrengthHint }}</p>
          </div>
          <div class="form-group" :class="{ 'has-error': !passwordMatch }">
            <label class="form-label"
                   for="vault-recoverpw-newpassword-repeat"
                   v-t="'vault.options.change_password.label.repeat'"></label>
            <input class="form-input"
                   type="password"
                   id="vault-recoverpw-newpassword-repeat"
                   v-model="newPasswordRepeat">
            <p class="form-input-hint"
               v-if="!passwordMatch"
               v-t="'vault.options.change_password.notmatch'"></p>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary float-right"
                :class="{ loading: $wait.is('resetting vault password') }"
                :disabled="!canResetPassword || $wait.is('resetting vault password')"
                @click="recoverPassword"
                v-t="'misc.done'"></button>
      </div>
    </div>
  </div>

</template>

<script>
import {mapGetters} from 'vuex'
import PasswordStrengthMeter from "@/components/PasswordStrengthMeter";

export default {
  name: "VaultPasswordRecoveryModal",
  components: {PasswordStrengthMeter},
  data: function() {
    return {
      masterkey: "",
      newPassword: "",
      newPasswordRepeat: "",
      passwordStrengthHint: ''
    }
  },
  computed: {
    ...mapGetters(['selectedVault']),
    passwordMatch() {
      return this.newPassword === this.newPasswordRepeat
    },
    masterkeyValid() {
      return this.masterkey.trim().length === 64
    },
    canResetPassword() {
      return this.masterkeyValid &&
          this.newPassword &&
          this.newPassword.length >= this.$store.getters.minimalPasswordLength &&
          this.passwordMatch
    }
  },
  methods: {
    close() {
      this.$emit('close')
    },
    showPasswordFeedback({warning}) {
      if (this.newPassword.length < this.$store.getters.minimalPasswordLength) {
        this.passwordStrengthHint = this.$t('misc.password.length_not_enough', {
          length: this.$store.getters.minimalPasswordLength
        })
        return
      }
      if (warning) {
        this.passwordStrengthHint = this.$t(`zxcvbn.${warning}`) || warning
      } else {
        this.passwordStrengthHint = ''
      }
    },
    recoverPassword() {
      // FIXME
    }
  },
  mounted() {
    this.$nextTick(() => this.$refs.masterkeyInput.focus())
  }
}
</script>

<style scoped>
.form-input-hint {
  margin-bottom: 0;
}
/deep/ .Password__strength-meter {
  margin: .4rem auto;
}
</style>