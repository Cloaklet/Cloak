<script setup lang="ts">
import PasswordStrengthMeter from "./PasswordStrengthMeter.vue";
import { computed, ref , watch} from 'vue';
import { useGlobalStore } from '@/stores/global';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
  using: 'password'|'masterkey',
}>();

const store = useGlobalStore();

const masterkey = ref('');
const password = ref('');
const newPassword = ref('');
const newPasswordRepeat = ref('');
const passwordStrengthHint = ref('');
// DOM refs
const passwordInput = ref();
const masterkeyInput = ref();
const emit = defineEmits(['close'])
const {t} = useI18n();

const passwordMatch = computed(() => newPassword.value === newPasswordRepeat.value);
const masterkeyValid = computed(() => masterkey.value.trim().length === 64);
const selectedVault = computed(() => store.selectedVault);
const canChangePassword = computed(() => {
  if (props.using === 'password') {
    return password.value &&
        newPassword.value &&
        newPassword.value.length >= store.minimalPasswordLength &&
        passwordMatch.value
  } else if (props.using === 'masterkey') {
    return masterkeyValid.value &&
        newPassword.value &&
        newPassword.value.length >= store.minimalPasswordLength &&
        passwordMatch.value
  }
  return false
})
const changeVaultPassword = () => {
  if (!selectedVault.value) {return}
  store.changeVaultPassword({
    vaultId: selectedVault.value.id,
    password: password.value,
    masterkey: masterkey.value,
    newpassword: newPassword.value
  }).then(() => emit('close'))
}
const showPasswordFeedback = ({warning}: {warning: string}) => {
  if (newPassword.value.length < store.minimalPasswordLength) {
    passwordStrengthHint.value = t('misc.password.length_not_enough', {
      length: store.minimalPasswordLength
    })
    return
  }
  if (warning) {
    passwordStrengthHint.value = t(`zxcvbn.${warning}`) || warning
  } else {
    passwordStrengthHint.value = ''
  }
}


watch(passwordInput, (newV, oldV) => {
  if (props.using === 'password' && typeof oldV === 'undefined' && newV) {
    newV.focus()
  }
})
watch (masterkeyInput, (newV, oldV) => {
  if (props.using === 'masterkey' && typeof oldV === 'undefined' && newV) {
    newV.focus()
  }
})

</script>
<template>
  <div class="modal active" @keydown.esc="$emit('close')">
    <a class="modal-overlay" aria-label="Close" @click="$emit('close')"></a>
    <div class="modal-container">
      <div class="modal-header">
        <a class="btn btn-clear float-right"
           aria-label="Close"
           @click="$emit('close')"></a>
        <div class="modal-title h5"
             v-if="using === 'password'"
             v-t="'vault.options.change_password.title'"></div>
        <div class="modal-title h5"
             v-if="using === 'masterkey'"
             v-t="'vault.options.recover_password.title'"></div>
      </div>
      <div class="modal-body">
        <div class="content">
          <div class="form-group"
               v-if="using === 'password'">
            <i18n-t tag="label"
                  class="form-label"
                  for="vault-chpw-oldpassword"
                  keypath="vault.options.change_password.label.password">
              <template #vaultname>{{ selectedVault?.name }}</template>
            </i18n-t>
            <input class="form-input"
                   type="password"
                   id="vault-chpw-oldpassword"
                   ref="passwordInput"
                   v-model="password">
          </div>
          <div class="form-group"
               v-if="using === 'masterkey'"
               :class="{ 'has-error': !masterkeyValid }">
            <i18n-t tag="label"
                  class="form-label"
                  for="vault-recoverpw-masterkey"
                  keypath="vault.options.recover_password.label.masterkey">
              <template #vaultname>{{ selectedVault?.name }}</template>
            </i18n-t>
            <input class="form-input"
                   type="text"
                   id="vault-recoverpw-masterkey"
                   ref="masterkeyInput"
                   v-model="masterkey">
          </div>
          <div class="form-group" :class="{ 'has-error': passwordStrengthHint }">
            <label class="form-label"
                   for="vault-chpw-newpassword"
                   v-t="'vault.options.change_password.label.newpassword'"></label>
            <PasswordStrengthMeter id="vault-chpw-newpassword"
                                   v-model="newPassword"
                                   :secure-length="store.minimalPasswordLength"
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
                   for="vault-chpw-newpassword-repeat"
                   v-t="'vault.options.change_password.label.repeat'"></label>
            <input class="form-input"
                   type="password"
                   id="vault-chpw-newpassword-repeat"
                   v-model="newPasswordRepeat">
            <p class="form-input-hint"
               v-if="!passwordMatch"
               v-t="'vault.options.change_password.notmatch'"></p>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary float-right"
                v-if="using === 'password'"
                :class="{ loading: false }"
                :disabled="!canChangePassword"
                @click="changeVaultPassword"
                v-t="'vault.options.change_password.button'"></button>
        <button class="btn btn-primary float-right"
                v-if="using === 'masterkey'"
                :class="{ loading: false}"
                :disabled="!canChangePassword "
                @click="changeVaultPassword"
                v-t="'misc.done'"></button>
      </div>
    </div>
  </div>

</template>


<style scoped>
.form-input-hint {
  margin-bottom: 0;
}
* :deep(.Password__strength-meter) {
  margin: .4rem auto;
}
</style>