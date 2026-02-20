// Debug HTMX syntax errors
if (window.htmx) {
  window.htmx.on("htmx:syntax:error", function (evt) {
    console.error("HTMX Syntax Error:", evt.detail);
    console.error("Element:", evt.detail.elt);
    console.error("Error message:", evt.detail.error);
  });
}
