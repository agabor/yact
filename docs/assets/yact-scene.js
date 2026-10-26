/*
 * yact-scene.js
 *
 * A small, dependency-free pixel-art renderer for the animation on the YACT landing page.
 * It draws into a 256x112 grayscale buffer, one frame at a time, entirely from code:
 * an agent robot walking to a ball in several steps, then a tiny truck that goes straight to it.
 * The renderer has no DOM access, so it also runs under Node (see module.exports below).
 */
(function (root) {
  'use strict';

  var W = 256, H = 112, GY = 94, BGC = 246, BX = 156, FPS = 24;
  var BAYER = [[0, 8, 2, 10], [12, 4, 14, 6], [3, 11, 1, 9], [15, 7, 13, 5]];

  // 5x7 pixel font: [row offset, rows]; glyphs are trimmed to their ink width when drawn
  var FONT = {
    "'": [0, ["#", "#"]],
    ".": [6, ["#"]],
    ":": [3, ["#", ".", "#"]],
    "C": [0, [".###.", "#...#", "#....", "#....", "#....", "#...#", ".###."]],
    "D": [0, ["####.", "#...#", "#...#", "#...#", "#...#", "#...#", "####."]],
    "E": [0, ["#####", "#....", "#....", "####.", "#....", "#....", "#####"]],
    "I": [0, [".###.", "..#..", "..#..", "..#..", "..#..", "..#..", ".###."]],
    "L": [0, ["#....", "#....", "#....", "#....", "#....", "#....", "#####"]],
    "N": [0, ["#...#", "##..#", "#.#.#", "#.#.#", "#..##", "#...#", "#...#"]],
    "P": [0, ["####.", "#...#", "#...#", "####.", "#....", "#....", "#...."]],
    "a": [2, [".###.", "....#", ".####", "#...#", ".####"]],
    "b": [0, ["#....", "#....", "####.", "#...#", "#...#", "#...#", "####."]],
    "c": [2, [".####", "#....", "#....", "#....", ".####"]],
    "d": [0, ["....#", "....#", ".####", "#...#", "#...#", "#...#", ".####"]],
    "e": [2, [".###.", "#...#", "#####", "#....", ".###."]],
    "f": [0, ["..##.", ".#...", "###..", ".#...", ".#...", ".#...", ".#..."]],
    "g": [2, [".####", "#...#", "#...#", "#...#", ".####", "....#", ".###."]],
    "h": [0, ["#....", "#....", "####.", "#...#", "#...#", "#...#", "#...#"]],
    "i": [0, ["..#..", ".....", ".##..", "..#..", "..#..", "..#..", ".###."]],
    "k": [0, ["#....", "#....", "#..#.", "#.#..", "##...", "#.#..", "#..#."]],
    "l": [0, [".##..", "..#..", "..#..", "..#..", "..#..", "..#..", ".###."]],
    "m": [2, ["##.#.", "#.#.#", "#.#.#", "#.#.#", "#.#.#"]],
    "n": [2, ["####.", "#...#", "#...#", "#...#", "#...#"]],
    "o": [2, [".###.", "#...#", "#...#", "#...#", ".###."]],
    "p": [2, ["####.", "#...#", "#...#", "#...#", "####.", "#....", "#...."]],
    "r": [2, ["#.##.", "##..#", "#....", "#....", "#...."]],
    "s": [2, [".####", "#....", ".###.", "....#", "####."]],
    "t": [1, [".#...", "####.", ".#...", ".#...", ".#...", "..##."]],
    "u": [2, ["#...#", "#...#", "#...#", "#..##", ".##.#"]],
    "v": [2, ["#...#", "#...#", "#...#", ".#.#.", "..#.."]],
    "x": [2, ["#...#", ".#.#.", "..#..", ".#.#.", "#...#"]],
    "y": [2, ["#...#", "#...#", "#...#", ".####", "....#", "....#", ".###."]]
  };

  // ---------------------------------------------------------------- helpers
  function pyround(x) {
    var f = Math.floor(x), d = x - f;
    if (d < 0.5) return f;
    if (d > 0.5) return f + 1;
    return f % 2 === 0 ? f : f + 1;
  }

  function smooth(u) {
    u = Math.min(1, Math.max(0, u));
    return u * u * (3 - 2 * u);
  }

  function lerp(a, b, s) { return a + (b - a) * s; }

  function put(img, x, y, v) {
    if (x >= 0 && x < W && y >= 0 && y < H) img[y * W + x] = v;
  }

  function fill(img, x, y, w, h, v) {
    var x1 = Math.max(x, 0), y1 = Math.max(y, 0);
    var x2 = Math.min(x + w, W), y2 = Math.min(y + h, H);
    for (var yy = y1; yy < y2; yy++) {
      for (var xx = x1; xx < x2; xx++) img[yy * W + xx] = v;
    }
  }

  // rectangle by height above the ground line: heights hb .. hb+hh-1 (height 0 is row GY-1)
  function hrect(img, x, hb, w, hh, v) {
    var y1 = GY - 1 - hb;
    fill(img, x, y1 - hh + 1, w, hh, v);
  }

  function thick(img, p0, p1, v) {
    var x0 = pyround(p0[0]), y0 = pyround(p0[1]);
    var x1 = pyround(p1[0]), y1 = pyround(p1[1]);
    var n = Math.max(Math.abs(x1 - x0), Math.abs(y1 - y0), 1);
    for (var i = 0; i <= n; i++) {
      fill(img, pyround(x0 + (x1 - x0) * i / n), pyround(y0 + (y1 - y0) * i / n), 2, 2, v);
    }
  }

  function mulberry32(seed) {
    var a = seed | 0;
    return function () {
      a = (a + 0x6D2B79F5) | 0;
      var t = Math.imul(a ^ (a >>> 15), 1 | a);
      t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
      return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };
  }

  // ---------------------------------------------------------------- pixel font
  var glyphCache = {};

  function ginfo(ch) {
    if (glyphCache[ch]) return glyphCache[ch];
    var off = FONT[ch][0], rows = FONT[ch][1];
    var c0 = 99, c1 = -1;
    rows.forEach(function (r) {
      for (var i = 0; i < r.length; i++) {
        if (r[i] === '#') { c0 = Math.min(c0, i); c1 = Math.max(c1, i); }
      }
    });
    glyphCache[ch] = { off: off, rows: rows, c0: c0, c1: c1 };
    return glyphCache[ch];
  }

  function textWidth(s, scale) {
    var w = 0;
    for (var k = 0; k < s.length; k++) {
      var ch = s[k];
      if (ch === ' ') { w += 4; continue; }
      var g = ginfo(ch);
      w += (g.c1 - g.c0 + 1) + 1;
    }
    return Math.max(0, w - 1) * scale;
  }

  function drawText(img, s, x, y, color, scale) {
    for (var k = 0; k < s.length; k++) {
      var ch = s[k];
      if (ch === ' ') { x += 4 * scale; continue; }
      var g = ginfo(ch);
      for (var r = 0; r < g.rows.length; r++) {
        var row = g.rows[r];
        for (var c = g.c0; c <= g.c1; c++) {
          if (c < row.length && row[c] === '#') {
            fill(img, x + (c - g.c0) * scale, y + (g.off + r) * scale, scale, scale, color);
          }
        }
      }
      x += (g.c1 - g.c0 + 2) * scale;
    }
  }

  function drawLine(img, partial, full, scale, y, color, cursor) {
    var x0 = Math.floor((W - textWidth(full, scale)) / 2);
    var pw = partial ? textWidth(partial, scale) : 0;
    drawText(img, partial, x0, y, color, scale);
    if (cursor) fill(img, x0 + pw + (partial ? scale : 0), y, 2 * scale, 7 * scale, color);
  }

  // ---------------------------------------------------------------- background
  function makeBg() {
    var rng = mulberry32(3);
    var img = new Uint8Array(W * H).fill(BGC);
    var x, y;
    for (y = 50; y < GY; y++) {
      var t = (y - 50) / (GY - 50);
      for (x = 0; x < W; x++) {
        if (t > BAYER[y % 4][x % 4] / 16) img[y * W + x] = 234;
      }
    }
    for (x = 0; x < W; x++) {
      var far = 9 * Math.sin(2 * Math.PI * 2 * x / W + 0.6) + 4 * Math.sin(2 * Math.PI * 5 * x / W + 1.9);
      for (y = pyround(77 - far); y < GY; y++) img[y * W + x] = 216;
    }
    for (x = 0; x < W; x++) {
      var near = 5 * Math.sin(2 * Math.PI * 3 * x / W + 2.2) + 2.5 * Math.sin(2 * Math.PI * 8 * x / W + 0.4);
      for (y = pyround(86 - near); y < GY; y++) img[y * W + x] = 200;
    }
    var cx = 226, cy = 24, r = 6;
    for (y = 0; y < H; y++) {
      for (x = 0; x < W; x++) {
        var d = (x - cx) * (x - cx) + (y - cy) * (y - cy);
        if (d <= r * r) img[y * W + x] = 214;
        else if (d <= r * r + 8 && d > r * r - 1) img[y * W + x] = 170;
      }
    }
    [[-2, -2], [2, 1], [0, 3], [-3, 2]].forEach(function (p) { img[(cy + p[1]) * W + cx + p[0]] = 180; });
    for (x = 0; x < W; x++) img[GY * W + x] = 70;
    for (y = GY + 1; y < H; y++) for (x = 0; x < W; x++) img[y * W + x] = 238;
    var specks = [70, 150, 190];
    for (var i = 0; i < 110; i++) {
      x = Math.floor(rng() * W);
      y = (GY + 3) + Math.floor(rng() * (H - (GY + 3)));
      img[y * W + x] = specks[Math.floor(rng() * 3)];
    }
    for (i = 0; i < 14; i++) {
      x = Math.floor(rng() * (W - 8));
      y = (GY + 4) + Math.floor(rng() * ((H - 1) - (GY + 4)));
      var len = 2 + Math.floor(rng() * 4);
      for (var dx = 0; dx < len; dx++) img[y * W + x + dx] = 100;
    }
    return img;
  }

  // ---------------------------------------------------------------- ball
  var BALL_MASK = ['01110', '11111', '11111', '11111', '01110'];

  function drawBall(img, x0, y0, ang) {
    for (var j = 0; j < 5; j++) {
      for (var i = 0; i < 5; i++) if (BALL_MASK[j][i] === '1') put(img, x0 + i, y0 + j, 120);
    }
    put(img, x0 + pyround(2 + 1.4 * Math.cos(ang)), y0 + pyround(2 + 1.4 * Math.sin(ang)), 236);
    put(img, x0 + pyround(2 - 1.4 * Math.cos(ang)), y0 + pyround(2 - 1.4 * Math.sin(ang)), 50);
  }

  function groundBall(img) {
    fill(img, BX - 2, GY + 1, 5, 1, 205);
    drawBall(img, BX - 2, GY - 1 - 2 - 2, 0.6);
  }

  // ---------------------------------------------------------------- robot
  var CHECK = ['....#', '...#.', '#.#..', '.#...'];

  function ik(s, h, L1, L2) {
    L1 = L1 || 8; L2 = L2 || 8;
    var sx = s[0], sy = s[1], hx = h[0], hy = h[1];
    var dx = hx - sx, dy = hy - sy;
    var d = Math.max(Math.hypot(dx, dy), 1e-3);
    var reach = L1 + L2 - 0.01;
    if (d > reach) {
      hx = sx + dx * reach / d; hy = sy + dy * reach / d;
      d = reach; dx = hx - sx; dy = hy - sy;
    }
    var a = (L1 * L1 - L2 * L2 + d * d) / (2 * d);
    var hh = Math.sqrt(Math.max(0, L1 * L1 - a * a));
    var ux = dx / d, uy = dy / d;
    var px = sx + a * ux, py = sy + a * uy;
    var e1 = [px - hh * uy, py + hh * ux];
    var e2 = [px + hh * uy, py - hh * ux];
    var sc = function (e) { return e[1] - 0.8 * e[0]; };
    return [sc(e1) >= sc(e2) ? e1 : e2, [hx, hy]];
  }

  function drawRobot(img, xr, st) {
    var xi = pyround(xr);
    var cd = pyround(st.cd || 0);
    var g = st.g || 0, amp = st.amp || 0;
    var bob = (amp > 0.3 && Math.cos(2 * g) > 0.4) ? 1 : 0;
    var bb = 9 - cd + bob;
    fill(img, xi - 9, GY + 1, 19, 1, 205);
    fill(img, xi - 6, GY + 2, 13, 1, 218);
    var shY = GY - 1 - (bb + 9);
    var bsx = xi - 6;
    var r = ik([bsx, shY], [bsx - 3 * amp * Math.sin(g), shY + 12]);
    var e = r[0], hh_ = r[1];
    thick(img, [bsx, shY], e, 125);
    thick(img, e, hh_, 125);
    fill(img, pyround(hh_[0]) - 1, pyround(hh_[1]), 3, 3, 95);
    [[-4, 0.0, 105], [3, Math.PI, 135]].forEach(function (leg) {
      var lx = leg[0], ph = leg[1], col = leg[2];
      var sw = pyround(3.5 * amp * Math.sin(g + ph));
      var lift = (amp > 0.3 && Math.cos(g + ph) > 0.5) ? 1 : 0;
      var x = xi + lx + sw;
      var ln = bb - (3 + lift);
      if (ln > 0) hrect(img, x - 2, 3 + lift, 4, ln, col);
      hrect(img, x - 2, lift, 7, 3, 60);
      hrect(img, x - 1, lift + 2, 4, 1, 105);
    });
    var bx = xi - 8;
    hrect(img, bx, bb, 16, 12, 60);
    hrect(img, bx + 1, bb + 1, 14, 10, 170);
    hrect(img, bx + 1, bb + 1, 14, 3, 135);
    hrect(img, bx + 1, bb + 10, 14, 1, 200);
    hrect(img, xi + 1, bb + 4, 6, 5, 90);
    hrect(img, xi + 2, bb + 5, 4, 3, 232);
    var led = st.led || 0;
    for (var k = 0; k < 3; k++) hrect(img, xi + 2 + k, bb + 6, 1, 1, k === led ? 40 : 190);
    hrect(img, bx + 2, bb + 9, 1, 1, 60);
    hrect(img, bx + 13, bb + 9, 1, 1, 60);
    var hdx = st.hdx || 0, hdy = st.hdy || 0;
    var hb0 = bb + 14 - hdy;
    hrect(img, xi - 2, bb + 12, 4, 2, 90);
    var hx0 = xi - 9 + hdx;
    hrect(img, hx0, hb0, 18, 13, 60);
    hrect(img, hx0 + 1, hb0 + 1, 16, 11, 178);
    hrect(img, hx0 + 1, hb0 + 11, 16, 1, 205);
    hrect(img, hx0 + 3, hb0 + 3, 12, 8, 95);
    hrect(img, hx0 + 4, hb0 + 4, 10, 6, 238);
    var eh = st.eye_h === undefined ? 3 : st.eye_h, edx = st.eye_dx || 0;
    var ey = hb0 + 5 + Math.floor((3 - eh) / 2);
    hrect(img, hx0 + 6 + edx, ey, 2, eh, 40);
    hrect(img, hx0 + 10 + edx, ey, 2, eh, 40);
    for (k = 0; k < 5; k++) hrect(img, hx0 + 5 + 2 * k, hb0 + 1, 1, 1, 100);
    hrect(img, hx0 - 3, hb0 + 4, 3, 6, 60);
    hrect(img, hx0 - 2, hb0 + 5, 1, 4, 125);
    hrect(img, hx0 + 18, hb0 + 4, 3, 6, 60);
    hrect(img, hx0 + 19, hb0 + 5, 1, 4, 125);
    var fsx = xi + 6;
    var rest = [fsx + 3 * amp * Math.sin(g), shY + 12];
    var tgt = st.hand, w = st.hw || 0;
    var hand = (!tgt || w <= 0) ? rest : [lerp(rest[0], tgt[0], w), lerp(rest[1], tgt[1], w)];
    r = ik([fsx, shY], hand);
    e = r[0];
    var h2 = r[1];
    thick(img, [fsx, shY], e, 85);
    thick(img, e, h2, 85);
    fill(img, fsx - 1, shY - 1, 3, 3, 55);
    fill(img, pyround(e[0]), pyround(e[1]), 2, 2, 55);
    var hx = pyround(h2[0]), hy = pyround(h2[1]);
    var ang = st.ball_ang;
    if (w > 0.5) {
      if (ang !== undefined && ang !== null) {
        drawBall(img, hx + 1, hy - 2, ang);
        var sc = st.scan;
        if (sc !== undefined && sc !== null && sc >= 0 && sc <= 4) {
          for (var i = 0; i < 5; i++) {
            if (BALL_MASK[sc][i] === '1') put(img, hx + 1 + i, hy - 2 + sc, 244);
          }
          put(img, hx + 7, hy - 2 + sc, 60);
        }
        if (st.check) {
          for (var j = 0; j < CHECK.length; j++) {
            for (i = 0; i < 5; i++) if (CHECK[j][i] === '#') put(img, hx + 2 + i, hy - 11 + j, 40);
          }
        }
      }
      fill(img, hx, hy - 3, 1, 7, 55);
      fill(img, hx, hy - 3, 4, 1, 55);
      fill(img, hx, hy + 3, 4, 1, 55);
    } else {
      fill(img, hx - 1, hy, 3, 3, 55);
    }
  }

  // ---------------------------------------------------------------- truck (wedge nose, open bed)
  var WEDGE = [[16, 3, 8], [17, 3, 7], [18, 3, 6], [19, 2, 5], [20, 2, 4], [21, 2, 3]];
  var WHEEL = ['.##.', '####', '####', '.##.'];
  var HUBS = [[1, 1], [2, 1], [2, 2], [1, 2]];

  function drawCarBody(img, L, ph) {
    var Li = pyround(L);
    var R = function (x, hb, w, hh, v) { hrect(img, Li + x, hb, w, hh, v); };
    fill(img, Li, GY + 1, 22, 1, 205);
    R(0, 3, 16, 3, 70);
    WEDGE.forEach(function (c) {
      R(c[0], c[1], 1, c[2] - c[1], 150);
      R(c[0], c[2] - 1, 1, 1, 70);
    });
    R(9, 6, 7, 3, 140);
    R(9, 9, 6, 2, 140);
    R(9, 11, 5, 1, 70);
    R(10, 8, 4, 3, 228);
    R(15, 7, 1, 1, 70);
    [2, 15].forEach(function (wx) {
      for (var j = 0; j < 4; j++) {
        for (var i = 0; i < 4; i++) if (WHEEL[j][i] === '#') put(img, Li + wx + i, GY - 1 - 3 + j, 40);
      }
      var hub = HUBS[ph % 4];
      put(img, Li + wx + hub[0], GY - 1 - 3 + hub[1], 205);
    });
    R(20, 2, 1, 1, 235);
  }

  function drawCarWall(img, L) {
    var Li = pyround(L);
    hrect(img, Li, 6, 9, 1, 135);
    hrect(img, Li, 7, 9, 1, 70);
    hrect(img, Li, 8, 1, 2, 70);
  }

  // ---------------------------------------------------------------- timelines
  function typed(full, fpc) {
    var n = Math.floor(full.length * fpc) + 1, out = [];
    for (var i = 0; i < n; i++) out.push(full.slice(0, Math.min(full.length, Math.floor(i / fpc))));
    return out;
  }

  function textTimeline(first, s1, y1, second, s2, y2, fpc1, hold, fpc2, endHold) {
    var tl = [];
    function add(p, full, sc, y) { tl.push([p, full, sc, y, Math.floor(tl.length / 8) % 2 === 0]); }
    var i;
    for (i = 0; i < 6; i++) add('', first, s1, y1);
    typed(first, fpc1).forEach(function (p) { add(p, first, s1, y1); });
    for (i = 0; i < hold; i++) add(first, first, s1, y1);
    var n = first.length;
    while (n > 0) { n = Math.max(0, n - 2); add(first.slice(0, n), first, s1, y1); }
    for (i = 0; i < 4; i++) add('', second, s2, y2);
    typed(second, fpc2).forEach(function (p) { add(p, second, s2, y2); });
    for (i = 0; i < endHold; i++) add(second, second, s2, y2);
    return tl;
  }

  // scene 2 has no prompt line: a second typed prompt would look like the loop restarting
  function introTimeline(title, sc, y, fpc, endHold) {
    var tl = [], i;
    function add(p, cur) { tl.push([p, title, sc, y, cur]); }
    for (i = 0; i < 10; i++) add('', false);
    typed(title, fpc).forEach(function (p) { add(p, Math.floor(tl.length / 8) % 2 === 0); });
    for (i = 0; i < endHold; i++) add(title, Math.floor(tl.length / 8) % 2 === 0);
    return tl;
  }

  var STATUS = ['Inspecting Context', 'Locating ball', 'Ensuring ball identity', 'Delivering ball'];

  function robotTimeline() {
    var V = 1.6, X0 = -24, XS = BX - 15, XE = 284;
    var A = Math.ceil((XS - X0) / V), Hn = Math.ceil((XE - XS) / V);
    var lens = [A, 10, 18, 8, 22, 60, 16, Hn, 24];
    var b = [0];
    lens.forEach(function (n) { b.push(b[b.length - 1] + n); });
    var grip = [BX - 3, GY - 3], exam = [XS + 13, 66];
    var out = [];
    for (var t = 0; t < b[b.length - 1]; t++) {
      var seg = 0;
      for (var i = 0; i < 9; i++) if (t >= b[i]) seg = i;
      var u = t - b[seg];
      var st = { g: 2 * Math.PI * t / 13.0, led: Math.floor(t / 5) % 3, eye_h: 3, eye_dx: 1 };
      var xr;
      if (seg === 0) { xr = X0 + V * u; st.amp = Math.min(1.0, u / 6); }
      else if (seg === 7) { xr = XS + V * u; st.amp = Math.min(1.0, u / 8); }
      else if (seg === 8) { xr = XE; }
      else { xr = XS; }
      if (seg === 1) st.amp = Math.max(0.0, 1 - u / lens[1]);
      var s;
      if (seg === 2) { st.cd = 5 * smooth(u / lens[2]); st.hand = grip; st.hw = smooth(u / lens[2]); }
      else if (seg === 3) { st.cd = 5; st.hand = grip; st.hw = 1.0; }
      else if (seg === 4) {
        s = smooth(u / lens[4]);
        st.cd = 5 * (1 - s);
        st.hand = [lerp(grip[0], exam[0], s), lerp(grip[1], exam[1], s)]; st.hw = 1.0;
      } else if (seg === 5) {
        st.hand = [exam[0], exam[1] + (Math.floor(u / 9) % 2 ? 1 : 0)]; st.hw = 1.0;
      } else if (seg === 6) {
        s = smooth(u / lens[6]);
        st.hand = [lerp(exam[0], XS + 10, s), lerp(exam[1], 86, s)]; st.hw = 1.0;
      } else if (seg === 7) {
        st.hand = [xr + 10, 86 + (Math.cos(st.g) > 0 ? 1 : 0)]; st.hw = 1.0;
      }
      if (t >= b[3] + 4) st.ball_ang = 0.42 * Math.max(0, Math.min(t, b[6]) - b[4]);
      if (seg === 0 && u < 55) st.eye_dx = pyround(Math.sin(t * 0.2));
      if (seg === 5) {
        st.eye_h = 2;
        st.eye_dx = [0, 1, 1, 0, -1, -1][Math.floor(u / 6) % 6];
        st.hdx = 1;
        st.hdy = Math.floor(u / 12) % 2 ? 1 : 0;
        st.scan = Math.floor(u / 2) % 12;
        st.check = u >= 40;
      }
      if ((t % 52 === 0 || t % 52 === 1) && seg !== 5) st.eye_h = 1;
      var sIdx;
      if (t < Math.floor(A * 0.45)) sIdx = 0;
      else if (t < b[4]) sIdx = 1;
      else if (t < b[6]) sIdx = 2;
      else sIdx = 3;
      out.push({ xr: xr, st: st, status: STATUS[sIdx], ballOnGround: t < b[3] + 4 });
    }
    return out;
  }

  // ---------------------------------------------------------------- the two scenes
  var PROMPT = 'Prompt: Pick up the ball.';
  var T1 = 'Coding agents';
  var T2 = 'yact.';
  var SUB = 'Not for vibe coding.';

  // ball flight for the truck: v px/frame, launch at the nose, lands in the bed (derived offline, see README of the page)
  var CAR_V = 3, CAR_NAPP = 52, FLIGHT_T = 20, FLIGHT_G = 0.33999999999999986;
  var FLIGHT_VY = (6 + 0.5 * FLIGHT_G * FLIGHT_T * FLIGHT_T) / FLIGHT_T;
  var FLIGHT_VB = CAR_V - 20.0 / FLIGHT_T;

  function buildScene1(bg) {
    var frames = [];
    textTimeline(PROMPT, 1, 16, T1, 2, 10, 1.5, 20, 1.5, 8).forEach(function (e) {
      frames.push(function (img) {
        groundBall(img);
        drawLine(img, e[0], e[1], e[2], e[3], 45, e[4]);
      });
    });
    robotTimeline().forEach(function (f) {
      frames.push(function (img) {
        if (f.ballOnGround) groundBall(img);
        if (f.xr < W + 30) drawRobot(img, f.xr, f.st);
        drawLine(img, T1, T1, 2, 10, 45, false);
        drawText(img, f.status, Math.floor((W - textWidth(f.status, 1)) / 2), 32, 80, 1);
      });
    });
    return frames;
  }

  function buildScene2(bg) {
    var frames = [];
    introTimeline(T2, 3, 6, 2.0, 6).forEach(function (e) {
      frames.push(function (img) {
        groundBall(img);
        drawLine(img, e[0], e[1], e[2], e[3], 45, e[4]);
      });
    });
    var hitL = BX - 24, L0 = hitL - CAR_V * CAR_NAPP;
    for (var k = 0; ; k++) {
      var L = L0 + CAR_V * k;
      if (L > W + 3) break;
      frames.push((function (k, L) {
        return function (img) {
          var tt = k - CAR_NAPP;
          if (tt < 0) groundBall(img);
          drawCarBody(img, L, k % 4);
          if (tt >= 0) {
            var bxp, byp;
            if (tt < FLIGHT_T) {
              var cx = BX + FLIGHT_VB * tt, h = 2 + FLIGHT_VY * tt - 0.5 * FLIGHT_G * tt * tt;
              bxp = cx; byp = GY - 1 - h;
              fill(img, pyround(cx) - 1, GY + 1, 3, 1, 208);
            } else {
              bxp = L + 4; byp = GY - 1 - 8;
            }
            drawBall(img, pyround(bxp) - 2, pyround(byp) - 2, 0.5 * Math.min(tt, FLIGHT_T));
            if (tt < 5) {
              [[0, 0], [-2, 1], [-3, -1]].forEach(function (p) {
                put(img, BX - 4 - tt + p[0], GY - 3 - Math.floor(tt / 2) + p[1], 198);
              });
            }
          }
          drawCarWall(img, L);
          drawLine(img, T2, T2, 3, 6, 45, false);
        };
      })(k, L));
    }
    var subs = typed(SUB, 1.2);
    for (var i = 0; i < 54; i++) subs.push(SUB);
    subs.forEach(function (p, i) {
      frames.push(function (img) {
        drawLine(img, T2, T2, 3, 6, 45, false);
        drawLine(img, p, SUB, 1, 40, 80, Math.floor(i / 8) % 2 === 0 && i < 45);
      });
    });
    return frames;
  }

  function fade(img, alpha) {
    for (var y = 0; y < H; y++) {
      for (var x = 0; x < W; x++) {
        if (BAYER[y % 4][x % 4] / 16 < alpha) img[y * W + x] = BGC;
      }
    }
  }

  /** Create the animation. render(i, out) fills `out` (Uint8Array, W*H, grayscale) with frame i. */
  function createScene() {
    var bg = makeBg();
    var s1 = buildScene1(bg), s2 = buildScene2(bg);
    var n = 8, total = s1.length + s2.length;

    function render(index, out) {
      var inFirst = index < s1.length;
      var local = inFirst ? index : index - s1.length;
      var len = inFirst ? s1.length : s2.length;
      var fn = inFirst ? s1[local] : s2[local];
      out.set(bg);
      fn(out);
      if (local < n) fade(out, 1 - (local + 1) / n);
      else if (local >= len - n) fade(out, (local - (len - n) + 1) / n);
      return out;
    }

    return {
      width: W,
      height: H,
      fps: FPS,
      frameCount: total,
      posterIndex: total - 12,
      render: render
    };
  }

  var api = { createScene: createScene, WIDTH: W, HEIGHT: H, FPS: FPS };
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  else root.YactScene = api;
})(typeof window !== 'undefined' ? window : this);
