<script setup lang="ts">
/*
Notice: this file is taken from
https://github.com/JokerQyou/vue-password-strength-meter/blob/13eec5dab7d1769ff156d11ad618861619cca4ae/src/components/PasswordStrengthMeter.vue
 */
import { computed, ref, watch } from 'vue';
import zxcvbn from 'zxcvbn'
const props = defineProps({
    /**
     * Binded value
     * @type {Object}
     */
    value: {
      type: String
    },
    /**
     * Password min length.
     * Right now only visual for the badge
     * @type {Number}
     */
    secureLength: {
      type: Number,
      default: 7
    },
    /**
     * Display badge:
     * The badge shows your
     * password character count
     * up to the defined secureLength
     * @type {Boolean}
     */
    badge: {
      type: Boolean,
      default: true
    },
    /**
     * Show password toggle:
     * Show icon to toggle
     * the password visibility
     */
    toggle: {
      type: Boolean,
      default: false
    },
    /**
     * Prop to toggle the
     * cleartext password if
     * toggle is disabled
     */
    showPassword: {
      type: Boolean,
      default: false
    },
    /**
     * Prop to change the
     * ref of the input
     */
    referenceValue: {
      type: String,
      default: 'input'
    },
    /**
     * Prop to toggle the
     * strength Meter if
     * User wants to implement
     * their own
     */
    showStrengthMeter: {
      type: Boolean,
      default: true
    },
    /**
     * Prop to toggle the
     * input element if
     * User wants to implement
     * their own input element
     */
    strengthMeterOnly: {
      type: Boolean,
      default: false
    },
    /**
     * CSS Class for the Input field
     * @type {String}
     */
    defaultClass: {
      type: String,
      default: 'Password__field'
    },
    /**
     * CSS Class for the disabled Input field
     * @type {String}
     */
    disabledClass: {
      type: String,
      default: 'Password__field--disabled'
    },
    /**
     * CSS Class for the badge
     * if a password does not match
     * the secureLength. Later for errors
     * @type {String}
     */
    errorClass: {
      type: String,
      default: 'Password__badge--error'
    },
    /**
     * CSS Class for the badge
     * if a password does match
     * the secureLength. Later for
     * success messages possible.
     * @type {String}
     */
    successClass: {
      type: String,
      default: 'Password__badge--success'
    },
    /**
     * CSS class for styling the
     * strength meter bars.
     * @type {String}
     */
    strengthMeterClass: {
      type: String,
      default: 'Password__strength-meter'
    },
    /**
     * strengthMeterFillClass sets the
     * individual strength width and fill
     * color of the strength meter bars.
     * @type {String}
     */
    strengthMeterFillClass: {
      type: String,
      default: 'Password__strength-meter--fill'
    },
    /**
     * Label for the show password icon
     */
    labelShow: {
      type: String,
      default: 'Show Password'
    },
    /**
     * Label for the hide password icon
     */
    labelHide: {
      type: String,
      default: 'Hide Password'
    },
    /**
     * @type String
     */
    userInputs: {
      type: Array<string>,
      default: () => []
    }
  });

  const emit = defineEmits(['hide', 'show', 'feedback', 'score', 'input', 'blur', 'focus'])
  const password = ref<string|null>(props.value||null);
  const shouldShowPassword = ref<boolean>(false);

  const passwordStrength = computed(() => password.value ? (
    zxcvbn(password.value, (props.userInputs.length >= 1 ? props.userInputs : undefined)).score
    ) : null);
    const isSecure = computed(() => password.value ? password.value.length >= props.secureLength : null);
    const isActive = computed(() => password.value && password.value.length > 0);
    const passwordCount = computed(() => password.value && (
      password.value.length > props.secureLength ? `${props.secureLength}+` : password.value.length
    ));
    const inputType = computed(() => shouldShowPassword.value || props.showPassword ? 'text' : 'password');
    const showPasswordLabel = computed(() => shouldShowPassword.value || props.showPassword ? props.labelHide : props.labelShow);

  const togglePassword = () => {
    if (shouldShowPassword.value) {
      emit('hide')
      shouldShowPassword.value = false
    } else {
      emit('show')
      shouldShowPassword.value = true
    }
  }
  const emitValue = (event: typeof emit.arguments[0], value: string) => {
    emit(event, value)
  }

  watch(password, (newValue) => {
    if (props.strengthMeterOnly) {
      emit('input', newValue)
    }
    emit('feedback', zxcvbn(newValue!).feedback)
  })
  watch(passwordStrength, (newValue) => {
    emit('score', newValue)
  })

</script>
<template>
  <div class="Password">
    <div
        v-if="!strengthMeterOnly"
        class="Password__group"
    >
      <input
          v-bind="$attrs"
          :type="inputType"
          :ref="referenceValue"
          :class="[defaultClass, $attrs.disabled ? disabledClass : '']"
          :value="value"
          @input="evt => emitValue('input', (evt.target as HTMLInputElement).value)"
          @blur="evt => emitValue('blur', (evt.target as HTMLInputElement).value)"
          @focus="evt => emitValue('focus', (evt.target as HTMLInputElement).value)"
      >
      <div class="Password__icons">
        <div
            v-if="badge"
            :class="[isSecure ? successClass : '', !isSecure && isActive ? errorClass : '' ]"
            class="Password__badge"
            v-cloak
        >
          {{ passwordCount }}
        </div>
        <div
            v-if="toggle"
            class="Password__toggle">
          <button
              type="button"
              class="btn-clean"
              :aria-label="showPasswordLabel"
              tabindex="-1"
              @click.prevent="togglePassword()">
            <svg v-if="shouldShowPassword" version="1.1" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
              <title>{{showPasswordLabel}}</title>
              <path d="M12 9c1.641 0 3 1.359 3 3s-1.359 3-3 3-3-1.359-3-3 1.359-3 3-3zM12 17.016c2.766 0 5.016-2.25 5.016-5.016s-2.25-5.016-5.016-5.016-5.016 2.25-5.016 5.016 2.25 5.016 5.016 5.016zM12 4.5c5.016 0 9.281 3.094 11.016 7.5-1.734 4.406-6 7.5-11.016 7.5s-9.281-3.094-11.016-7.5c1.734-4.406 6-7.5 11.016-7.5z"></path>
            </svg>
            <svg v-else version="1.1" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
              <title>{{showPasswordLabel}}</title>
              <path d="M11.859 9h0.141c1.641 0 3 1.359 3 3v0.188zM7.547 9.797c-0.328 0.656-0.563 1.406-0.563 2.203 0 2.766 2.25 5.016 5.016 5.016 0.797 0 1.547-0.234 2.203-0.563l-1.547-1.547c-0.188 0.047-0.422 0.094-0.656 0.094-1.641 0-3-1.359-3-3 0-0.234 0.047-0.469 0.094-0.656zM2.016 4.266l1.266-1.266 17.719 17.719-1.266 1.266c-1.124-1.11-2.256-2.213-3.375-3.328-1.359 0.563-2.813 0.844-4.359 0.844-5.016 0-9.281-3.094-11.016-7.5 0.797-1.969 2.109-3.656 3.75-4.969-0.914-0.914-1.812-1.844-2.719-2.766zM12 6.984c-0.656 0-1.266 0.141-1.828 0.375l-2.156-2.156c1.219-0.469 2.578-0.703 3.984-0.703 5.016 0 9.234 3.094 10.969 7.5-0.75 1.875-1.922 3.469-3.422 4.734l-2.906-2.906c0.234-0.563 0.375-1.172 0.375-1.828 0-2.766-2.25-5.016-5.016-5.016z"></path>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div v-if="showStrengthMeter" :class="[strengthMeterClass]">
      <div :class="[strengthMeterFillClass]" :data-score="passwordStrength"></div>
    </div>
  </div>
</template>


<style lang="scss">
[v-cloak] {
  display: none;
}

.Password {
  margin:  0 auto;
}

.Password__group {
  position: relative;
}

.Password__strength-meter {
  position: relative;
  height: 3px;
  background: #DDD;
  margin: 10px auto 20px;
  border-radius: 3px;
}

.Password__strength-meter:before, .Password__strength-meter:after {
  content: '';
  height: inherit;
  background: transparent;
  display: block;
  border-color: #FFF;
  border-style: solid;
  border-width: 0 5px 0 5px;
  position: absolute;
  width: 20%;
  z-index: 10;
}

.Password__strength-meter:before {
  left: 20%;
}

.Password__strength-meter:after {
  right: 20%;
}

.Password__strength-meter--fill {
  background: transparent;
  height: inherit;
  position: absolute;
  width: 0;
  border-radius: inherit;
  transition: width 0.5s ease-in-out, background 0.25s;
}

.Password__strength-meter--fill[data-score='0'] {
  background: darkred;
  width: 20%;
}

.Password__strength-meter--fill[data-score='1'] {
  background: orangered;
  width: 40%;
}

.Password__strength-meter--fill[data-score='2'] {
  background: orange;
  width: 60%;
}

.Password__strength-meter--fill[data-score='3'] {
  background: yellowgreen;
  width: 80%;
}

.Password__strength-meter--fill[data-score='4'] {
  background: green;
  width: 100%;
}

.Password__field {
  background-color: #f1f1f1;
  border: 1px solid #f1f1f1;
  border-radius: 2px;
  box-sizing: border-box;
  font-size: 14px;
  padding: 13px;
  width: 100%;
}

.Password__field--disabled {
  background-color: #f6f6f6;
  border: 1px solid #f6f6f6;
}

.Password__icons {
  position: absolute;
  top: 0;
  right: 0;
  height: 100%;
  display: flex;
  flex-direction: row;
  justify-content: flex-end;
  align-items: center;
}

.Password__toggle {
  line-height: 1.1;
  margin-right: 13px;
}

.Password__badge {
  position: relative;
  color: white;
  border-radius: 6px;
  padding: 3px;
  width: 30px;
  height: 15px;
  font-size: 14px;
  line-height: 1.1;
  margin-right: 13px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.Password__badge--error {
  background: red;
}

.Password__badge--success {
  background: #1bbf1b;
}

.btn-clean {
  appearance: none;
  background: none;
  border: none;
  cursor: pointer;
  outline: none;
  color: #777777;
  padding: 0;

  svg {
    fill: currentColor;
  }

  &:hover, &:focus {
    color: #404B69;
  }
}

</style>
