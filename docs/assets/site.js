/* Adds a copy button to every code block marked data-copy. The pages read fine without it. */
(function () {
  'use strict';
  if (!navigator.clipboard) return;
  document.querySelectorAll('pre[data-copy]').forEach(function (pre) {
    var button = document.createElement('button');
    button.type = 'button';
    button.className = 'copy';
    button.textContent = 'Copy';
    button.addEventListener('click', function () {
      var code = pre.querySelector('code');
      navigator.clipboard.writeText(code.innerText.replace(/\n$/, '')).then(function () {
        button.textContent = 'Copied';
        setTimeout(function () { button.textContent = 'Copy'; }, 1600);
      });
    });
    pre.appendChild(button);
  });
})();
