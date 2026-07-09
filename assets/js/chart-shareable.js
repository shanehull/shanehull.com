/**
 * Chart Shareable — Syncs chart control state with URL query params.
 * On load with URL params: overrides HTML defaults, triggers chart reload,
 *   and prevents the HTML load trigger from firing with stale defaults.
 * On user change: writes current control state to the browser URL.
 */
(function () {
  var CONTROLS_SELECTOR = ".chart-controls input";
  var suppressLoad = false;

  function getControls() {
    return document.querySelectorAll(CONTROLS_SELECTOR);
  }

  function readURL() {
    var params = new URLSearchParams(window.location.search);
    if (window.location.search.length === 0) return false;

    getControls().forEach(function (input) {
      if (input.type === "radio") {
        input.checked = params.get(input.name) === input.value;
      } else if (input.type === "checkbox") {
        input.checked = params.get(input.name) === "on";
      }
    });

    return true;
  }

  function writeURL() {
    var url = new URL(window.location);

    getControls().forEach(function (input) {
      if (input.type === "radio") {
        if (input.checked) url.searchParams.set(input.name, input.value);
      } else if (input.type === "checkbox") {
        if (input.checked) {
          url.searchParams.set(input.name, "on");
        } else {
          url.searchParams.set(input.name, "off");
        }
      }
    });

    window.history.replaceState({}, "", url);
  }

  document.body.addEventListener("htmx:beforeRequest", function (evt) {
    if (!suppressLoad) return;
    if (
      evt.detail.requestConfig &&
      evt.detail.requestConfig.trigger &&
      evt.detail.requestConfig.trigger === "load"
    ) {
      evt.preventDefault();
      suppressLoad = false;
    }
  });

  function init() {
    var controls = getControls();
    if (controls.length === 0) return;

    if (readURL()) {
      suppressLoad = true;
      var radio = document.querySelector(
        '.chart-controls input[type="radio"]:checked'
      );
      if (radio && window.htmx) htmx.trigger(radio, "change");
    }

    controls.forEach(function (input) {
      input.addEventListener("change", writeURL);
    });

    window.addEventListener("popstate", function () {
      if (readURL()) {
        var radio = document.querySelector(
          '.chart-controls input[type="radio"]:checked'
        );
        if (radio && window.htmx) htmx.trigger(radio, "change");
      }
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
