<script setup lang="ts">
import { useGlobalStore } from '@/stores/global';
import { computed, onMounted, ref, nextTick} from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
  title?: string,
  mode: 'file'|'directory',
  okBtn?: string,
}>();

type fileItem = {
  name: string
  type: 'file'|'directory'
}

const {t} = useI18n();
const store = useGlobalStore();
const pwd = ref('');
const items = ref<fileItem[]>([]);
const selected = ref<fileItem|null>(null);
const sep = ref('');
const emit = defineEmits(['selected', 'close']);

const hasParent = computed(() => pwd.value && pwd.value !== '/');
const canFinishSelection = computed(() => props.mode === 'directory' || (props.mode === 'file' && !!selected.value));

const listSubPaths = (d: string) => {
  store.listSubPaths({path: d}).then(data => {
    pwd.value = data.pwd
    sep.value = data.sep
    items.value = data.items
  })
}
const clickOnItem = (item: fileItem) => {
  // Click on a directory enters it
  if (item.type === 'directory') {
    selected.value = null;
    return listSubPaths(`${pwd.value}${sep.value}${item.name}`)
  }
  // Click on a file selects it, but only if we're in file mode
  // Only allow selecting gocryptfs.conf
  if (props.mode === 'file' && item.name === 'gocryptfs.conf') {
    selected.value = item
  }
}
const clickOnParentPath = () => {
  selected.value = null;
  listSubPaths(`${pwd.value}${sep.value}..`)
}
const selectItem = () => {
  if (props.mode === 'file') {
    emit('selected', `${pwd.value}${sep.value}${selected.value!.name}`)
  } else if (props.mode === 'directory') {
    emit('selected', `${pwd.value}`)
  }
}

onMounted(() => {
  nextTick(()=>{
    listSubPaths('$HOME');
  })
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
        <div class="modal-title h5">{{ title || t('select.default_title') }}</div>
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
                @click="selectItem">{{ okBtn || t('misc.select') }}</button>
        <button class="btn ml-1" aria-label="Close" @click="$emit('close')" v-t="'misc.cancel'"></button>
      </div>
    </div>
  </div>
</template>


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