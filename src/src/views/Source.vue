<template>
  <div style="min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center">
    <div class="card">
      <a class="back" href="/">&larr; 返回首页</a>
      <h2>源码</h2>
      <p style="color:var(--text2);font-size:14px;margin-bottom:16px">Waterproof Wall 源代码托管在 Gitee</p>
      <a class="btn" :href="repoUrl" target="_blank">访问 Gitee 仓库 &rarr;</a>
      <div class="url">{{ repoUrl }}</div>
    </div>
  </div>
</template>

<script setup>
const repoUrl = 'https://gitee.com/waterproof-wall/waterproof-wall'
</script>

<style scoped>
.card { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 24px; width: min(460px,90vw); text-align: center; }
.card h2 { font-size: 18px; margin-bottom: 8px; }
.btn { display: inline-flex; align-items: center; justify-content: center; height: 40px; padding: 0 20px; border: 0; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer; background: var(--primary); color: #fff; text-decoration: none; }
.url { margin-top: 12px; padding: 10px; background: var(--input-bg); border: 1px solid var(--border); border-radius: 6px; font-size: 13px; color: var(--text2); word-break: break-all; }
.back { display: inline-block; margin-bottom: 16px; color: var(--primary); font-size: 14px; }
</style>
