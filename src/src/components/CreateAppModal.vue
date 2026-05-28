<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal">
      <h3>创建新应用</h3>
      <div class="form-group"><label>应用名称</label><input v-model="name" placeholder="如：官网登录页"></div>
      <div class="form-group">
        <label>允许域名（每行一个）</label>
        <textarea v-model="domains" placeholder="example.com&#10;localhost"></textarea>
        <div class="form-hint">留空不限制</div>
      </div>
      <div class="btn-group">
        <button class="btn btn-ghost" @click="$emit('close')">取消</button>
        <button class="btn btn-primary" @click="submit">创建</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['close', 'created'])
const name = ref('')
const domains = ref('')

function submit() {
  if (!name.value.trim()) return
  const domainList = domains.value.trim().split('\n').map(d => d.trim()).filter(Boolean)
  emit('created', { name: name.value.trim(), domains: domainList })
}
</script>