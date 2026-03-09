function debounce(func, wait) {
  let timeout;
  return function (...args) {
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(this, args), wait);
  };
}

/**
 * Check if an input is visible (not hidden by display:none or similar)
 * Handles both the input itself and parent containers
 */
function isInputVisible(input) {
  // If offsetParent is null, the element is hidden (including hidden parents)
  if (input.offsetParent === null) return false;

  // Also check computed style up the parent chain for display: none
  let element = input;
  while (element) {
    const style = window.getComputedStyle(element);
    if (style.display === "none") return false;
    element = element.parentElement;
  }

  return true;
}

// Global flag to prevent URL updates during initialization
let shareableInitializing = false;
let shareableObserver = null;

function makeShareable(calculateCallback) {
  // Disable ALL URL updates until initialization is completely done
  shareableInitializing = true;

  // Disconnect old observer from previous page navigation
  if (shareableObserver) {
    shareableObserver.disconnect();
    shareableObserver = null;
  }

  const params = new URLSearchParams(window.location.search);
  let shouldCalc = false;

  const inputs = document.querySelectorAll("main input, main select");

  if (inputs.length === 0) {
    shareableInitializing = false;
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
    // Skip URL updates during initialization
    if (shareableInitializing) return;

    const url = new URL(window.location);

    // Get fresh list of all inputs (includes dynamically added ones)
    const allInputs = document.querySelectorAll("main input, main select");

    allInputs.forEach((input) => {
      // Skip inputs marked with data-no-share
      if (input.hasAttribute("data-no-share")) return;

      if (!input.id) return;

      // Check visibility: only encode visible inputs
      if (!isInputVisible(input)) {
        url.searchParams.delete(input.id);
        return;
      }

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

  // Function to attach listeners to inputs
  const attachListeners = (inputList) => {
    inputList.forEach((input) => {
      // Skip if already has listeners (check via a marker)
      if (input.dataset.shareableAttached) return;

      input.addEventListener("input", debouncedUpdate);
      input.addEventListener("change", updateURL);
      input.dataset.shareableAttached = "true";
    });
  };

  attachListeners(inputs);

  // Watch for DOM changes: new elements, visibility changes, etc.
  shareableObserver = new MutationObserver((mutations) => {
    // Skip all observer processing during initialization
    if (shareableInitializing) return;

    let needsUpdate = false;

    mutations.forEach((mutation) => {
      // New nodes added: attach listeners but don't update URL yet
      if (mutation.type === "childList") {
        const newInputs = mutation.addedNodes;
        newInputs.forEach((node) => {
          if (node.nodeType === 1) {
            // Element node
            // Check if the node itself is an input or contains inputs
            const nodeInputs = node.querySelectorAll
              ? [node, ...node.querySelectorAll("input, select")]
              : [];

            nodeInputs.forEach((el) => {
              if (
                (el.tagName === "INPUT" || el.tagName === "SELECT") &&
                !el.hasAttribute("data-no-share")
              ) {
                attachListeners([el]);
                // Don't set needsUpdate here - only update on user interaction
              }
            });
          }
        });
      }

      // Attribute changes (visibility/display changes): these are user actions
      if (
        mutation.type === "attributes" &&
        (mutation.attributeName === "style" ||
          mutation.attributeName === "class")
      ) {
        needsUpdate = true;
      }
    });

    // Update URL only on visibility changes, not on new element additions
    if (needsUpdate) {
      updateURL();
    }
  });

  // Done initializing - allow URL updates from here on
  // Delay observer setup until next event loop to avoid triggering during init
  setTimeout(() => {
    // Observe main for all changes
    const main = document.querySelector("main");
    if (main && shareableObserver) {
      shareableObserver.observe(main, {
        childList: true,
        subtree: true,
        attributes: true,
        attributeFilter: ["style", "class"],
      });
    }
    shareableInitializing = false;
  }, 0);
}
