<template>
  <div class="docs-page">
    <h1>接入文档</h1>

    <section>
      <h2>验证流程</h2>
      <div class="flow">
        <div class="flow-step">
          <div class="flow-num">1</div>
          <div><strong>前端</strong>调用 <code>POST /challenge</code> 获取验证码图片和参数</div>
        </div>
        <div class="flow-step">
          <div class="flow-num">2</div>
          <div><strong>用户</strong>拖动滑块完成拼图，前端收集轨迹</div>
        </div>
        <div class="flow-step">
          <div class="flow-num">3</div>
          <div><strong>前端</strong>调用 <code>POST /verify</code> 提交轨迹，获得 ticket</div>
        </div>
        <div class="flow-step">
          <div class="flow-num">4</div>
          <div><strong>后端</strong>调用 <code>POST /ticket/check</code> 验证 ticket 有效性</div>
        </div>
      </div>
      <p>以下接口公开访问，无需用户认证，但需提供 <code>appId</code>。</p>
    </section>

    <section>
      <h2>前端接入</h2>
      <p>使用 SDK 一行代码即可完成前端接入，无需手动调用 challenge 和 verify 接口。</p>
      <div class="api-block">
        <h4 style="margin:0 0 12px;font-size:15px">1. 引入 SDK</h4>
        <pre class="api-code">&lt;script src="https://007.hallo.run/sdk/captcha.js"&gt;&lt;/script&gt;</pre>
      </div>
      <div class="api-block">
        <h4 style="margin:0 0 12px;font-size:15px">2. 初始化并验证</h4>
        <pre class="api-code">const captcha = new SliderCaptcha({
  appId: '你的AppId',
  endpoint: 'https://007.hallo.run',
  scene: 'default',   // 可选，业务场景
  bizId: '',          // 可选，业务标识
  timeout: 10000      // 可选，超时时间（ms），默认 10s
});

captcha.verify().then(result => {
  // result.success === true 时验证通过
  // result.ticket 为验证凭据，传给后端验票
  console.log(result.ticket);
}).catch(err => {
  // 验证失败或超时
  console.error(err.message);
});</pre>
      </div>
      <div class="api-block">
        <h4 style="margin:0 0 12px;font-size:15px">3. 后端验票</h4>
        <p style="margin-bottom:12px">将前端返回的 <code>result.ticket</code> 发送到你的后端，后端调用 <code>POST /api/v1/ticket/check</code> 验证。</p>
        <pre class="api-code">// 前端：验证成功后将 ticket 发给后端
fetch('/your-api/verify-captcha', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    ticket: result.ticket,
    scene: 'default',
    bizId: ''
  })
});</pre>
      </div>
      <div class="api-block">
        <h4 style="margin:0 0 12px;font-size:15px">返回值说明</h4>
        <pre class="api-code">// 验证成功 — verify() resolve，result 即下方对象
{
  "success": true,
  "ticket": "eyJhbGci...",
  "expiresIn": 180,
  "score": 0.85
}

// 验证失败 — SDK 自动提示"验证失败，请重试"，不触发 then 也不触发 catch，
// 返回值示例（仅供后端接口参考，SDK 不会将此值暴露给调用方）：
{
  "success": false,
  "reason": "position_mismatch",
  "score": 0
}

// 用户取消 — catch 收到 Error: "captcha cancelled"
// 超时 — catch 收到 Error: "captcha timeout"</pre>
      </div>
    </section>

    <section>
      <h2>后端验票</h2>
      <p>后端调用此接口验证前端传来的 ticket。ticket 一次性有效，验证后即刻失效。需要 HMAC 签名认证。</p>

      <div class="api-block">
        <h3><span class="method post">POST</span> https://007.hallo.run/api/v1/ticket/check</h3>

        <h4 style="margin:20px 0 8px;font-size:14px;color:var(--text2)">签名计算</h4>
        <pre class="api-code">signingInput = METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + NONCE + "\n" + BODY
signature = Base64URL(HMAC-SHA256(appSecret, signingInput))</pre>
        <p style="margin-top:8px">其中 METHOD 为 <code>POST</code>，PATH 为 <code>/api/v1/ticket/check</code>，TIMESTAMP 为 Unix 时间戳（秒），NONCE 为随机字符串，BODY 为请求原始 JSON 字符串。</p>

        <div class="api-detail">
          <div class="api-label">请求头</div>
          <pre class="api-code">Content-Type: application/json
X-Captcha-App-Key: 你的AppKey
X-Captcha-Timestamp: 1700000000
X-Captcha-Nonce: 随机字符串
X-Captcha-Signature: Base64URL(HMAC-SHA256签名)</pre>
        </div>
        <div class="api-detail">
          <div class="api-label">请求体</div>
          <pre class="api-code">{ "appId": "app_xxx", "ticket": "eyJ...", "scene": "default", "bizId": "" }</pre>
        </div>
        <div class="api-detail">
          <div class="api-label success">成功响应 (200)</div>
          <pre class="api-code">{
  "success": true,
  "appId": "app_xxx",
  "captchaId": "cha_xxx",
  "scene": "default",
  "bizId": "",
  "issuedAt": 1700000000,
  "expiresAt": 1700000180
}</pre>
        </div>
        <div class="api-detail">
          <div class="api-label error">错误码</div>
          <pre class="api-code">401  missing_app_key          缺少 X-Captcha-App-Key 头
401  app_not_found             appKey 对应的应用不存在
401  invalid_signature         HMAC 签名验证失败
401  invalid_ticket            ticket 格式错误或已过期
401  ticket_used               ticket 已被使用
401  ticket_context_mismatch   ticket 中的 appId/scene/bizId 不匹配
403  app_disabled              应用已禁用</pre>
        </div>
      </div>
    </section>

    <section>
      <h2>Go 完整示例</h2>
      <pre class="api-code lang-go">package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "math/rand"
    "net/http"
    "strings"
    "time"
)

const (
    appKey    = "你的AppKey"
    appSecret = "你的AppSecret"
    endpoint  = "https://007.hallo.run"
)

// CheckTicket 验证前端传来的 ticket
func CheckTicket(ticket, scene, bizID string) error {
    body, _ := json.Marshal(map[string]string{
        "appId":  "",
        "ticket": ticket,
        "scene":  scene,
        "bizId":  bizID,
    })

    ts := fmt.Sprintf("%d", time.Now().Unix())
    nonce := fmt.Sprintf("n_%d_%d", time.Now().UnixNano(), rand.Int63())
    method := "POST"
    path := "/api/v1/ticket/check"

    // 构造签名输入
    signingInput := strings.Join([]string{method, path, ts, nonce, string(body)}, "\n")

    // HMAC-SHA256 签名
    mac := hmac.New(sha256.New, []byte(appSecret))
    mac.Write([]byte(signingInput))
    signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

    // 发送请求
    req, _ := http.NewRequest(method, endpoint+path, strings.NewReader(string(body)))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Captcha-App-Key", appKey)
    req.Header.Set("X-Captcha-Timestamp", ts)
    req.Header.Set("X-Captcha-Nonce", nonce)
    req.Header.Set("X-Captcha-Signature", signature)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    data, _ := io.ReadAll(resp.Body)
    var result struct {
        Success bool   `json:"success"`
        Error   string `json:"error"`
    }
    json.Unmarshal(data, &result)
    if !result.Success {
        return fmt.Errorf("ticket invalid: %s", result.Error)
    }
    return nil
}

func main() {
    err := CheckTicket("前端传来的ticket", "default", "")
    if err != nil {
        fmt.Println("验票失败:", err)
    } else {
        fmt.Println("验票成功")
    }
}</pre>
    </section>
  </div>
</template>

<script setup>
</script>

<style scoped>
.docs-page {
  max-width: 880px;
  margin: 0 auto;
  padding: 48px 32px 80px;
}
.docs-page h1 {
  font-size: 32px;
  font-weight: 800;
  margin-bottom: 32px;
}
.docs-page h2 {
  font-size: 22px;
  font-weight: 700;
  margin: 48px 0 20px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border);
}
.docs-page h3 {
  font-size: 16px;
  font-weight: 700;
  margin: 0 0 12px;
}
.docs-page p {
  font-size: 14px;
  color: var(--text2);
  margin-bottom: 8px;
  line-height: 1.7;
}
.docs-page code {
  background: var(--input-bg);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  color: #f87171;
}
.flow {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 20px 0 32px;
}
.flow-step {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 14px;
  color: var(--text2);
}
.flow-num {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
}
.api-block {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 24px;
  margin-bottom: 20px;
}
.api-detail {
  margin-top: 16px;
}
.api-label {
  font-size: 13px;
  font-weight: 700;
  margin-bottom: 6px;
  color: var(--text2);
}
.api-label.success { color: #4ade80; }
.api-label.error { color: #f87171; }
.api-code {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 14px 18px;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre;
  color: #e2e8f0;
  font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
}
.api-code.lang-go {
  font-size: 12.5px;
  line-height: 1.55;
}
.method {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 700;
  margin-right: 8px;
  vertical-align: middle;
}
.method.post {
  background: #22c55e22;
  color: #4ade80;
}
</style>