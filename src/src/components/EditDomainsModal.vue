<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal">
      <h3>编辑允许域名</h3>
      <div class="form-hint" style="margin-bottom:10px">每行一个域名，留空不限制</div>
      <div class="form-group"><textarea v-model="domains" style="height:120px" placeholder="example.com&#10;localhost"></textarea></div>
      <div class="btn-group">
        <button class="btn btn-ghost" @click="$emit('close')">取消</button>
        <button class="btn btn-primary" @click="submit">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({ initial: { type: String, default: '' } })
const emit = defineEmits(['close', 'save'])
const domains = ref(props.initial)

function submit() {
  const list = domains.value.trim().split('\n').map(d => d.trim()).filter(Boolean)
  emit('save', list)
}
</script>