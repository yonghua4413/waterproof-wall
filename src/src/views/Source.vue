<template>
  <div style="min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center">
    <div class="card">
      <a class="back" href="/">&larr; 返回首页</a>
      <h2>源码</h2>
      <p style="color:var(--text2);font-size:14px;margin-bottom:16px">Waterproof Wall 源代码托管在以下仓库</p>
      <div class="repo-list">
        <a class="btn" :href="giteeUrl" target="_blank">Gitee &rarr;</a>
        <a class="btn" :href="githubUrl" target="_blank">GitHub &rarr;</a>
      </div>
      <div class="url">{{ giteeUrl }}</div>
      <div class="url">{{ githubUrl }}</div>
    </div>
  </div>
</template>

<script setup>
const giteeUrl = 'https://gitee.com/yonghua4413/waterproof-wall'
const githubUrl = 'https://github.com/yonghua4413/waterproof-wall'
</script>

<style scoped>
.card { background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 24px; width: min(460px,90vw); text-align: center; }
.card h2 { font-size: 18px; margin-bottom: 8px; }
.btn { display: inline-flex; align-items: center; justify-content: center; height: 40px; padding: 0 20px; border: 0; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer; background: var(--primary); color: #fff; text-decoration: none; }
.url { margin-top: 12px; padding: 10px; background: var(--input-bg); border: 1px solid var(--border); border-radius: 6px; font-size: 13px; color: var(--text2); word-break: break-all; }
.repo-list { display: flex; gap: 12px; justify-content: center; margin-bottom: 12px; }
.back { display: inline-block; margin-bottom: 16px; color: var(--primary); font-size: 14px; }
</style>
