<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal">
      <h3>应用密钥</h3>
      <div class="form-hint" style="margin-bottom:10px;color:var(--danger)">App Secret 仅创建和重置时展示一次，请妥善保存！</div>
      <div class="field-row"><label>App ID</label><div class="field-val">{{ app.appId }}</div></div>
      <div class="field-row"><label>App Key</label><div class="field-val">{{ app.appKey }}</div></div>
      <div class="field-row"><label>App Secret</label><div class="field-val secret">{{ app.appSecret }}</div></div>
      <div class="btn-group"><button class="btn btn-primary" @click="$emit('close')">知道了</button></div>
    </div>
  </div>
</template>

<script setup>
defineProps({ app: { type: Object, default: () => ({}) } })
defineEmits(['close'])
</script>