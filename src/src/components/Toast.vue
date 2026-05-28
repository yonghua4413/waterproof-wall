<template>
  <div class="toast" :class="[toastState.type, { show: toastState.showing }]" v-if="toastState.showing || toastState.message">{{ toastState.message }}</div>
</template>

<script setup>
import { toastState } from '../composables/useToast'
</script>