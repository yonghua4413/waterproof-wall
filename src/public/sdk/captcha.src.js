(function () {
  class SliderCaptcha {
    constructor(options) {
      this.endpoint = (options && options.endpoint ? options.endpoint : "").replace(/\/$/, "");
      this.appId = options && options.appId ? options.appId : "";
      this.scene = options && options.scene ? options.scene : "default";
      this.bizId = options && options.bizId ? options.bizId : "";
      this.timeout = options && options.timeout ? options.timeout : 10000;
    }

    async verify() {
      const challenge = await this.requestChallenge();
      return this.open(challenge);
    }

    async requestChallenge() {
      const res = await fetch(this.endpoint + "/api/v1/challenge", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({appId: this.appId, scene: this.scene, bizId: this.bizId})
      });
      if (!res.ok) throw new Error("captcha challenge failed");
      return res.json();
    }

    open(challenge) {
      return new Promise((resolve, reject) => {
        const ui = createUI(challenge);
        const track = [];
        let downTime = 0;
        let dragging = false;
        let startX = 0;
        let lastX = 0;

        const cleanup = () => {
          document.removeEventListener("pointermove", onMove);
          document.removeEventListener("pointerup", onUp);
          window.clearTimeout(timer);
          ui.root.remove();
        };

        const fail = (message) => {
          ui.message.textContent = message || "验证失败，请重试";
          ui.slider.style.transform = "translateX(0px)";
          ui.piece.style.transform = "translateX(0px)";
          track.length = 0;
          dragging = false;
        };

        const onDown = (ev) => {
          dragging = true;
          startX = ev.clientX;
          lastX = 0;
          track.length = 0;
          downTime = performance.now();
          track.push({x: 0, y: 0, t: 0});
          ui.message.textContent = "";
          ui.handle.setPointerCapture(ev.pointerId);
        };

        const onMove = (ev) => {
          if (!dragging) return;
          const max = challenge.width - 46;
          const x = Math.max(0, Math.min(max, Math.round(ev.clientX - startX)));
          const y = Math.round(ev.clientY - ui.stage.getBoundingClientRect().top - challenge.pieceY);
          lastX = x;
          ui.slider.style.transform = "translateX(" + x + "px)";
          ui.piece.style.transform = "translateX(" + x + "px)";
          track.push({x, y, t: Math.round(performance.now() - downTime)});
        };

        const onUp = async () => {
          if (!dragging) return;
          dragging = false;
          track.push({x: lastX, y: 0, t: Math.round(performance.now() - downTime)});
          try {
            const res = await fetch(this.endpoint + "/api/v1/verify", {
              method: "POST",
              headers: {"Content-Type": "application/json"},
              body: JSON.stringify({
                appId: this.appId,
                captchaId: challenge.captchaId,
                scene: this.scene,
                bizId: this.bizId,
                x: lastX,
                track
              })
            });
            const data = await res.json();
            if (!res.ok || !data.success) {
              fail("验证失败，请重试");
              return;
            }
            window.clearTimeout(timer);
            ui.stage.style.display = "none";
            ui.bar.style.display = "none";
            ui.message.style.display = "none";
            ui.check.style.display = "flex";
            window.setTimeout(() => {
              cleanup();
              resolve(data);
            }, 800);
          } catch (err) {
            cleanup();
            reject(err);
          }
        };

        ui.close.addEventListener("click", () => {
          cleanup();
          reject(new Error("captcha cancelled"));
        });
        ui.handle.addEventListener("pointerdown", onDown);
        document.addEventListener("pointermove", onMove);
        document.addEventListener("pointerup", onUp);
        document.body.appendChild(ui.root);

        const timer = window.setTimeout(() => {
          if (document.body.contains(ui.root)) {
            cleanup();
            reject(new Error("captcha timeout"));
          }
        }, this.timeout);
      });
    }
  }

  function createUI(challenge) {
    injectCSS();
    const root = el("div", "sc-mask");
    const panel = el("div", "sc-panel");
    const top = el("div", "sc-top");
    const title = el("div", "sc-title", "安全验证");
    const close = el("button", "sc-close", "x");
    const stage = el("div", "sc-stage");
    const bg = el("img", "sc-bg");
    const piece = el("img", "sc-piece");
    const bar = el("div", "sc-bar");
    const text = el("div", "sc-bar-text", "拖动滑块完成拼图");
    const slider = el("div", "sc-slider");
    const handle = el("button", "sc-handle", ">");
    const message = el("div", "sc-message");
    const check = el("div", "sc-check");
    check.innerHTML = '<svg viewBox="0 0 48 48"><circle cx="24" cy="24" r="22" class="sc-check-circle"/><path d="M14 24 l6 6 l14-14" class="sc-check-path"/></svg>';

    panel.style.width = challenge.width + "px";
    stage.style.width = challenge.width + "px";
    stage.style.height = challenge.height + "px";
    bg.src = challenge.background;
    piece.src = challenge.piece;
    piece.style.top = challenge.pieceY + "px";
    bar.style.width = challenge.width + "px";
    slider.appendChild(handle);
    top.appendChild(title);
    top.appendChild(close);
    stage.appendChild(bg);
    stage.appendChild(piece);
    bar.appendChild(text);
    bar.appendChild(slider);
    panel.appendChild(top);
    panel.appendChild(stage);
    panel.appendChild(bar);
    panel.appendChild(message);
    panel.appendChild(check);
    root.appendChild(panel);
    return {root, close, stage, piece, bar, slider, handle, message, check};
  }

  function el(tag, className, text) {
    const node = document.createElement(tag);
    node.className = className;
    if (text) node.textContent = text;
    return node;
  }

  function injectCSS() {
    if (document.getElementById("slider-captcha-css")) return;
    const style = document.createElement("style");
    style.id = "slider-captcha-css";
    style.textContent = `
.sc-mask{position:fixed;inset:0;z-index:2147483647;display:flex;align-items:center;justify-content:center;background:rgba(14,18,24,.45);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
.sc-panel{position:relative;background:#fff;border:1px solid #d9dee7;border-radius:8px;box-shadow:0 16px 50px rgba(0,0,0,.22);padding:14px;box-sizing:content-box}
.sc-top{height:26px;display:flex;align-items:center;justify-content:space-between;margin-bottom:10px}
.sc-title{font-size:15px;font-weight:650;color:#172033}
.sc-close{width:24px;height:24px;border:0;background:#f1f3f6;color:#596273;border-radius:50%;cursor:pointer}
.sc-stage{position:relative;overflow:hidden;background:#eef2f7;border-radius:6px;user-select:none;touch-action:none}
.sc-bg{position:absolute;inset:0;width:100%;height:100%;display:block}
.sc-piece{position:absolute;left:0;width:58px;height:58px;filter:drop-shadow(0 4px 8px rgba(0,0,0,.22));will-change:transform}
.sc-bar{position:relative;height:42px;margin-top:12px;background:#f2f5f9;border:1px solid #d8dee8;border-radius:4px;overflow:hidden;user-select:none;touch-action:none}
.sc-bar-text{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;font-size:14px;color:#6b7280}
.sc-slider{position:absolute;left:0;top:0;height:42px;width:46px;will-change:transform}
.sc-handle{width:46px;height:42px;border:0;background:#1f7ae0;color:#fff;font-size:22px;font-weight:700;cursor:pointer}
.sc-message{height:20px;margin-top:8px;font-size:13px;color:#d34836}
.sc-check{display:none;position:absolute;top:40px;left:0;right:0;bottom:0;align-items:center;justify-content:center}
.sc-check svg{width:52px;height:52px}
.sc-check-circle{fill:none;stroke:#16a34a;stroke-width:3;stroke-dasharray:138.23;stroke-dashoffset:138.23;animation:sc-check-draw .5s ease forwards}
.sc-check-path{fill:none;stroke:#16a34a;stroke-width:3;stroke-linecap:round;stroke-linejoin:round;stroke-dasharray:26;stroke-dashoffset:26;animation:sc-check-draw .3s .35s ease forwards}
@keyframes sc-check-draw{to{stroke-dashoffset:0}}
`;
    document.head.appendChild(style);
  }

  window.SliderCaptcha = SliderCaptcha;
})();
