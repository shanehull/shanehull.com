/**
 * back-link.js - pagination-aware "back to tools" links.
 *
 * The tools list stores its own path in sessionStorage; back-links rewrite
 * their href to that saved path so "← Tools" returns to the exact list page
 * the visitor came from, not always page one. Falls back to "/tools/" when no
 * tools-list visit has happened in this tab. Works under hx-boost because the
 * link stays a real anchor with the correct href.
 */
(function () {
  var TOOLS_LIST = /^\/tools(\/page\/\d+)?[\/]?$/;
  var KEY = 'tools-list-path';

  function remember() {
    if (TOOLS_LIST.test(window.location.pathname)) {
      try {
        sessionStorage.setItem(KEY, window.location.pathname);
      } catch (e) {}
    }
  }

  function apply() {
    var saved;
    try {
      saved = sessionStorage.getItem(KEY);
    } catch (e) {}
    if (!saved) return;

    var links = document.querySelectorAll('a.back-link');
    Array.prototype.forEach.call(links, function (a) {
      if (a.getAttribute('href') === '/tools/') {
        a.setAttribute('href', saved);
      }
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () {
      remember();
      apply();
    });
  } else {
    remember();
    apply();
  }

  // hx-boost swaps the body without reloading; re-run after each swap.
  document.body.addEventListener('htmx:afterSwap', function () {
    remember();
    apply();
  });
})();