function debounce(func, wait) {
  let timeout;
  return function (...args) {
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(this, args), wait);
  };
}

function makeShareable(calculateCallback) {
  const params = new URLSearchParams(window.location.search);
  let shouldCalc = false;

  const inputs = document.querySelectorAll("main input, main select");

  if (inputs.length === 0) {
    return;
  }

  inputs.forEach((input) => {
    // Skip inputs marked with data-no-share
    if (input.hasAttribute("data-no-share")) return;

    if (input.id && params.has(input.id)) {
      const val = params.get(input.id);
      if (input.type === "checkbox" || input.type === "radio") {
        if (input.type === "radio") {
          if (input.value === val) input.checked = true;
        } else {
          input.checked = val === "true";
        }
      } else {
        input.value = val;
      }
      shouldCalc = true;
    }
  });

  if (shouldCalc && typeof calculateCallback === "function") {
    calculateCallback();
  }

  const updateURL = () => {
    const url = new URL(window.location);

    inputs.forEach((input) => {
      // Skip inputs marked with data-no-share
      if (input.hasAttribute("data-no-share")) return;

      if (!input.id) return;

      if (input.type === "checkbox") {
        if (input.checked) url.searchParams.set(input.id, "true");
        else url.searchParams.delete(input.id);
      } else if (input.type === "radio") {
        if (input.checked) url.searchParams.set(input.name, input.value);
      } else {
        if (input.value) url.searchParams.set(input.id, input.value);
        else url.searchParams.delete(input.id);
      }
    });

    window.history.replaceState({}, "", url);
  };

  const debouncedUpdate = debounce(updateURL, 300);

  inputs.forEach((input) => {
    input.addEventListener("input", debouncedUpdate);
    input.addEventListener("change", updateURL);
  });
}
