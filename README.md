# Waterproof Wall · 滑动验证码平台

自托管多租户滑动验证码平台，HMAC 签名验票，一行代码接入。

## 特性

- **多租户隔离** — 每个应用独立 appId / appKey / appSecret
- **HMAC-SHA256 签名验票** — 防伪造、防重放
- **App Secret 加密存储** — 仅创建/重置时明文返回一次
- **行为轨迹分析** — 滑动速度、加速度区分人机
- **用量统计** — 按日统计请求数、通过率、验票数
- **Vue 3 + Vite 前端** — 单页应用，nginx 静态部署

## 项目结构

```
.
├── cmd/api/          # API 服务（Go），监听 :8088
├── cmd/web/          # 前端构建输出（npm run build 生成）
├── src/              # 前端源码（Vue 3 + Vite）
│   ├── src/          # Vue 组件、页面、composables
│   ├── public/sdk/   # SDK 源文件（captcha.src.js）
│   └── .env          # VITE_APP_DOMAIN 配置
├── internal/         # Go 后端代码
├── deploy/           # 部署配置
│   ├── nginx.conf    # Nginx 反向代理配置
│   └── 007.service   # systemd 服务文件
├── config.example.yml
└── go.mod
```

## 快速启动

### 1. 配置

```bash
cp config.example.yml config.yml
# 编辑 config.yml，设置 secret、db.password 等
# 敏感参数建议通过环境变量覆盖：
export CAPTCHA_SECRET="your-32-byte-random-secret-here"
export CAPTCHA_DB_PASSWORD="your-db-password"
```

### 2. 启动 API

```bash
go run ./cmd/api/ -c config.yml
```

API 默认监听 `:8088`。

### 3. 前端开发

```bash
cd src
npm install
npm run dev      # 开发模式，访问 http://localhost:5173
```

### 4. 前端构建

```bash
cd src
npm run build    # 输出到 cmd/web/
```

修改域名：编辑 `src/.env` 中的 `VITE_APP_DOMAIN`，然后重新 build。

### 5. 生产部署

#### 交叉编译（Linux/CentOS）

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o captcha-api ./cmd/api/
```

#### 部署文件

| 文件 | 目标位置 |
| --- | --- |
| `captcha-api` | `/opt/007/captcha-api` |
| `cmd/web/` | `/opt/007/web/` |
| `config.yml` | `/opt/007/config.yml` |
| `deploy/nginx.conf` | `/etc/nginx/conf.d/007.hallo.run.conf` |
| `deploy/007.service` | `/etc/systemd/system/007.service` |

```bash
sudo systemctl enable 007
sudo systemctl start 007
sudo nginx -t && sudo systemctl reload nginx
```

## 域名配置

前端域名为 `VITE_APP_DOMAIN`（默认 `007.hallo.run`）：

- **首页** `https://007.hallo.run/` — 介绍 + 控制台
- **示例** `https://007.hallo.run/example` — SDK 演示
- **SDK** `https://007.hallo.run/sdk/captcha.js` — 压缩版
- **API** `https://007.hallo.run/api/` — 反向代理到 `:8088`

Nginx 配置：`/api/` 转发到后端，其余 `try_files` 回退 `index.html`。

## 接入流程

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  浏览器    │     │ 验证码服务 │     │ 业务后端  │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │  POST /challenge (appId)  │          │
     │───────────────────────►│          │
     │  背景图 + 滑块图       │          │
     │◄───────────────────────│          │
     │  POST /verify (appId, track)     │
     │───────────────────────►│          │
     │  ticket               │          │
     │◄───────────────────────│          │
     │  提交表单 + ticket     │          │
     │─────────────────────────────────►│
     │                       │  POST /ticket/check (appKey + HMAC签名)
     │                       │◄─────────│
     │                       │  验票结果  │
     │                       │──────────►│
     │  业务结果              │          │
     │◄──────────────────────────────────│
```

### 前端接入

```html
<script src="https://007.hallo.run/sdk/captcha.js"></script>
<script>
const captcha = new SliderCaptcha({
  appId: 'your-app-id',
  endpoint: 'https://007.hallo.run/api'
});
captcha.verify().then(result => {
  // result.ticket 提交给业务后端校验
});
</script>
```

### 后端验票（Go）

```go
signingInput := "POST\n/api/v1/ticket/check\n" + timestamp + "\n" + nonce + "\n" + body
mac := hmac.New(sha256.New, []byte(appSecret))
mac.Write([]byte(signingInput))
signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

// 请求头：
// X-Captcha-App-Key: {appKey}
// X-Captcha-Timestamp: {timestamp}
// X-Captcha-Nonce: {nonce}
// X-Captcha-Signature: {signature}
```

### 后端验票（Node.js）

```js
const signingInput = ["POST", "/api/v1/ticket/check", ts, nonce, body].join("\n");
const signature = createHmac("sha256", appSecret).update(signingInput).digest("base64url");

// 请求头同上
```

## HTTP API

### 平台管理

| 方法 | 路径 | 说明 | 认证 |
| --- | --- | --- | --- |
| POST | `/api/v1/users/register` | 用户注册 | 无 |
| POST | `/api/v1/users/login` | 用户登录 | 无 |
| POST | `/api/v1/apps` | 创建应用 | Bearer Token |
| GET | `/api/v1/apps` | 列出应用 | Bearer Token |
| POST | `/api/v1/apps/rotate-secret` | 重置 Secret | Bearer Token |
| POST | `/api/v1/apps/update-domains` | 更新域名 | Bearer Token |
| GET | `/api/v1/apps/usage?date=` | 调用统计 | Bearer Token |

### 验证码

| 方法 | 路径 | 说明 | 认证 |
| --- | --- | --- | --- |
| POST | `/api/v1/challenge` | 生成挑战 | appId |
| POST | `/api/v1/verify` | 提交验证 | appId |
| POST | `/api/v1/ticket/check` | 校验票据 | appKey + HMAC |

## 配置

通过 `config.yml` 加载，敏感参数支持 `CAPTCHA_` 前缀环境变量覆盖：

| 环境变量 | 说明 |
| --- | --- |
| `CAPTCHA_SECRET` | HMAC 签名密钥（生产必须替换） |
| `CAPTCHA_ADDR` | 监听地址，默认 `:8088` |
| `CAPTCHA_DB_PASSWORD` | MySQL 密码 |
| `CAPTCHA_PLATFORM_STORE` | `mysql` 或 `memory` |

详见 [config.example.yml](config.example.yml)。

## 数据库

MySQL 自动建表：

| 表 | 说明 |
| --- | --- |
| `users` | 注册用户 |
| `apps` | 应用（appId / appKey / secret_cipher / domains） |
| `captcha_events` | 验证事件日志 |

appSecret 使用 AES-GCM 加密存储。

## 安全建议

- `config.yml` 不入版本库（已在 `.gitignore`）
- 生产必须替换 `secret`，32 字节以上随机密钥
- appSecret 只保管在业务后端，绝不泄露到前端
- `/api/v1/ticket/check` 建议在网关层限制只允许业务后端 IP 访问
- 多实例部署使用 `CAPTCHA_STORE=redis`
- 票据校验成功后立即执行业务，不要缓存结果