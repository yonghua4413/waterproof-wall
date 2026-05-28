<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal">
      <h3>{{ title }}</h3>
      <p style="font-size:14px;color:var(--text2);margin-bottom:8px">{{ message }}</p>
      <div class="btn-group">
        <button class="btn btn-ghost" @click="$emit('close')">取消</button>
        <button class="btn btn-danger" @click="$emit('confirm')">确认</button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({ title: { type: String, default: '确认' }, message: { type: String, default: '' } })
defineEmits(['close', 'confirm'])
</script>