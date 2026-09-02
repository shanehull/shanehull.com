/**
 * KaTeX auto-render bootstrap.
 * Renders LaTeX in the page body after load and after each HTMX swap
 * (including hx-boost navigation). CSP compliant: no inline scripts.
 */
(function () {
  function render() {
    if (typeof renderMathInElement === "function") {
      renderMathInElement(document.body, {
        delimiters: [
          { left: "$$", right: "$$", display: true },
          { left: "\\[", right: "\\]", display: true },
          { left: "\\(", right: "\\)", display: false },
        ],
        throwOnError: false,
      });
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", render);
  } else {
    render();
  }

  document.body.addEventListener("htmx:afterSwap", render);
})();
