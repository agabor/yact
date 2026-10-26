/*
 * site.js: runs the hero animation (see yact-scene.js) and adds copy buttons to code blocks.
 * The pages are fully readable without it.
 */
(function () {
  'use strict';

  function setupScene() {
    var canvas = document.getElementById('scene');
    var wrap = document.getElementById('scene-wrap');
    var toggle = document.getElementById('scene-toggle');
    if (!canvas || !wrap || !window.YactScene) return;

    var scene = window.YactScene.createScene();
    var W = scene.width, H = scene.height;
    var ctx = canvas.getContext('2d');
    var image = ctx.createImageData(W, H);
    var gray = new Uint8Array(W * H);

    function draw(index) {
      scene.render(index, gray);
      var d = image.data;
      for (var p = 0, q = 0; p < gray.length; p++, q += 4) {
        d[q] = d[q + 1] = d[q + 2] = gray[p];
        d[q + 3] = 255;
      }
      ctx.putImageData(image, 0, 0);
    }

    // On wide screens every art pixel covers a whole number of device pixels, so pixels stay evenly
    // sized. On narrow screens the scene fills the width so the small text stays readable.
    function fit() {
      var available = wrap.clientWidth - 4;
      var dpr = window.devicePixelRatio || 1;
      var scale = available < 600 ? available / W : Math.floor(available * dpr / W) / dpr;
      canvas.style.width = Math.round(W * scale) + 'px';
      canvas.style.height = Math.round(H * scale) + 'px';
    }

    var reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
    var userPaused = reduced.matches;
    var onScreen = true;
    var frame = userPaused ? scene.posterIndex : 0;
    var last = 0, acc = 0, raf = 0;

    function running() { return !userPaused && onScreen && !document.hidden; }

    function tick(now) {
      raf = 0;
      if (!running()) return;
      if (!last) last = now;
      acc += now - last;
      last = now;
      var step = 1000 / scene.fps;
      var n = Math.floor(acc / step);
      if (n > 0) {
        acc -= n * step;
        frame = (frame + n) % scene.frameCount;
        draw(frame);
      }
      raf = window.requestAnimationFrame(tick);
    }

    function schedule() {
      if (!raf && running()) {
        last = 0;
        raf = window.requestAnimationFrame(tick);
      }
    }

    function label() {
      if (toggle) toggle.textContent = userPaused ? 'Play animation' : 'Pause animation';
    }

    if (toggle) {
      toggle.hidden = false;
      toggle.addEventListener('click', function () {
        userPaused = !userPaused;
        label();
        schedule();
      });
    }
    if (reduced.addEventListener) {
      reduced.addEventListener('change', function () {
        userPaused = reduced.matches;
        if (userPaused) { frame = scene.posterIndex; draw(frame); }
        label();
        schedule();
      });
    }
    document.addEventListener('visibilitychange', schedule);
    if ('IntersectionObserver' in window) {
      new IntersectionObserver(function (entries) {
        onScreen = entries[0].isIntersecting;
        schedule();
      }).observe(canvas);
    }
    if ('ResizeObserver' in window) new ResizeObserver(fit).observe(wrap);
    else window.addEventListener('resize', fit);

    document.documentElement.classList.add('scene-ready');
    fit();
    draw(frame);
    label();
    schedule();
  }

  function setupCopyButtons() {
    if (!navigator.clipboard || !navigator.clipboard.writeText) return;
    var blocks = document.querySelectorAll('pre[data-copy]');
    Array.prototype.forEach.call(blocks, function (pre) {
      var holder = document.createElement('div');
      holder.className = 'code';
      pre.parentNode.insertBefore(holder, pre);
      holder.appendChild(pre);
      var button = document.createElement('button');
      button.type = 'button';
      button.className = 'copy';
      button.textContent = 'Copy';
      button.addEventListener('click', function () {
        var source = pre.querySelector('code') || pre;
        navigator.clipboard.writeText(source.innerText.replace(/\s+$/, '')).then(function () {
          button.textContent = 'Copied';
          window.setTimeout(function () { button.textContent = 'Copy'; }, 1600);
        });
      });
      holder.appendChild(button);
    });
  }

  setupScene();
  setupCopyButtons();
})();
