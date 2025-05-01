// ==UserScript==
// @name         Binance Multi-Target Price Monitor (Compact UI)
// @namespace    http://tampermonkey.net/
// @version      0.9
// @description  用紧凑的自定义表单设置多目标价格和误差，WebAudio API合成beep提醒。
// @match        https://www.binance.com/*
// @grant        GM_setValue
// @grant        GM_getValue
// @grant        GM_deleteValue
// @grant        GM_registerMenuCommand
// ==/UserScript==

(function() {
    'use strict';
  
    let timer = null;
    let lastAlertTimes = {};
    let audioCtx, oscillator, gainNode;
  
    function initAudio() {
      if (!audioCtx) audioCtx = new (window.AudioContext || window.webkitAudioContext)();
    }
  
    function playBeep() {
      stopBeep();
      oscillator = audioCtx.createOscillator();
      gainNode   = audioCtx.createGain();
      oscillator.type = 'sine';
      oscillator.frequency.setValueAtTime(440, audioCtx.currentTime);
      gainNode.gain.setValueAtTime(0.2, audioCtx.currentTime);
      oscillator.connect(gainNode).connect(audioCtx.destination);
      oscillator.start();
    }
  
    function stopBeep() {
      if (oscillator) {
        oscillator.stop();
        oscillator.disconnect();
        gainNode.disconnect();
        oscillator = null;
        gainNode   = null;
      }
    }
  
    function getPrice() {
      const el = document.querySelector('.t-headline1');
      if (!el) return null;
      const num = parseFloat(el.textContent.replace(/,/g,'').trim());
      return isNaN(num) ? null : num;
    }
  
    function showForm() {
      document.querySelector('#__priceMonitorForm')?.remove();
      const saved = JSON.parse(GM_getValue('monitorSettings', '{}'));
      const savedTargets = saved.targets || '';
      const savedTol     = saved.tolerance || '';
  
      const overlay = document.createElement('div');
      overlay.id = '__priceMonitorForm';
      Object.assign(overlay.style, {
        position: 'fixed',
        top: '10px',
        right: '10px',
        zIndex: 10000,
        background: '#fff',
        border: '1px solid #ccc',
        borderRadius: '6px',
        padding: '8px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.2)',
        width: '200px',
        fontFamily: 'Arial, sans-serif',
        fontSize: '13px',
        color: '#333'
      });
  
      overlay.innerHTML = `
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
          <strong style="color:#2EBD85; font-size:14px;">监控设置</strong>
          <span id="__pmClose" style="cursor:pointer; font-size:14px;">✕</span>
        </div>
        <label style="display:block; margin-bottom:4px;">目标价格（逗号分隔）
          <input id="__pmTargets" style="width:100%; padding:4px; margin-top:2px; border:1px solid #ccc; border-radius:4px; font-size:12px;" value="${savedTargets}">
        </label>
        <label style="display:block; margin-bottom:6px;">误差范围 (tolerance)
          <input id="__pmTol" style="width:100%; padding:4px; margin-top:2px; border:1px solid #ccc; border-radius:4px; font-size:12px;" placeholder="留空自动" value="${savedTol}">
        </label>
        <div style="text-align:right;">
          <button id="__pmStart" style="padding:4px 8px; background:#2EBD85; color:#fff; border:none; border-radius:4px; font-size:12px; cursor:pointer;">启动</button>
        </div>
      `;
      document.body.appendChild(overlay);
  
      overlay.querySelector('#__pmClose').onclick = () => overlay.remove();
      overlay.querySelector('#__pmStart').onclick = () => {
        const tInput = overlay.querySelector('#__pmTargets').value.trim();
        const tolInput = overlay.querySelector('#__pmTol').value.trim();
        if (!tInput) { alert('请输入目标价格'); return; }
        const targets = tInput.split(',').map(s => parseFloat(s.trim())).filter(n => !isNaN(n));
        if (!targets.length) { alert('价格格式有误'); return; }
        let tol;
        if (tolInput) {
          tol = parseFloat(tolInput);
          if (isNaN(tol) || tol < 0) { alert('误差格式有误'); return; }
        } else {
          const ds = targets.map(n => (n.toString().split('.')[1]||'').length);
          const d = Math.max(...ds);
          tol = Math.pow(10, -d);
        }
        GM_setValue('monitorSettings', JSON.stringify({ targets: tInput, tolerance: tolInput }));
        overlay.remove();
        startMonitor(targets, tol);
      };
    }
  
    async function startMonitor(targets, tolerance) {
      if (timer) clearInterval(timer);
      document.querySelector('#__stopMonitorBtn')?.remove();
      stopBeep();
      lastAlertTimes = {};
  
      initAudio();
      if (audioCtx.state === 'suspended') await audioCtx.resume();
  
      const btn = document.createElement('button');
      btn.id = '__stopMonitorBtn';
      btn.textContent = '停止';
      Object.assign(btn.style, {
        position:'fixed', top:'10px', right:'220px', zIndex:10000,
        padding:'6px 10px', background:'#f44336', color:'#fff',
        border:'none', borderRadius:'4px', cursor:'pointer', fontSize:'12px'
      });
      btn.onclick = () => {
        clearInterval(timer);
        stopBeep();
        btn.remove();
      };
      document.body.appendChild(btn);
  
      timer = setInterval(() => {
        const price = getPrice();
        const now = Date.now();
        if (price === null) return;
        targets.forEach(t => {
          if (Math.abs(price - t) <= tolerance) {
            const last = lastAlertTimes[t] || 0;
            if (now - last > 60000) {
              console.log(`监控触发！目标 ${t}，当前 ${price}`);
              playBeep();
              lastAlertTimes[t] = now;
            }
          }
        });
      }, 2000);
    }
  
    GM_registerMenuCommand('开始监控', showForm);
  })();