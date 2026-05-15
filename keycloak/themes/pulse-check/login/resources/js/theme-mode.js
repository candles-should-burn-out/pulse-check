(() => {
  const storageKey = "pulse-check-theme-mode";

  const isThemeMode = (mode) => mode === "dark" || mode === "light";

  const getCookieMode = () => {
    const cookie = document.cookie
      .split("; ")
      .find((entry) => entry.startsWith(`${storageKey}=`));
    const mode = cookie?.split("=")[1] ?? null;

    return isThemeMode(mode) ? mode : null;
  };

  const getStoredMode = () => {
    const cookieMode = getCookieMode();

    if (cookieMode) {
      return cookieMode;
    }

    try {
      const mode = localStorage.getItem(storageKey);
      return isThemeMode(mode) ? mode : "light";
    } catch {
      return "light";
    }
  };

  const applyMode = (mode, toggle) => {
    document.documentElement.dataset.theme = mode;
    document.documentElement.style.colorScheme = mode;

    if (!toggle) {
      return;
    }

    toggle.setAttribute("aria-pressed", mode === "dark" ? "true" : "false");
    toggle.title =
      mode === "dark" ? "Включить светлую тему" : "Включить темную тему";
  };

  const saveMode = (mode) => {
    try {
      localStorage.setItem(storageKey, mode);
    } catch {
      // Theme selection is still applied for the current page.
    }

    document.cookie = `${storageKey}=${mode}; path=/; max-age=31536000; SameSite=Lax`;
  };

  applyMode(getStoredMode());

  document.addEventListener("DOMContentLoaded", () => {
    const toggle = document.querySelector(".theme-toggle");

    if (!(toggle instanceof HTMLButtonElement)) {
      return;
    }

    applyMode(
      document.documentElement.dataset.theme === "dark" ? "dark" : "light",
      toggle
    );

    toggle.addEventListener("click", () => {
      const nextMode =
        document.documentElement.dataset.theme === "dark" ? "light" : "dark";

      applyMode(nextMode, toggle);
      saveMode(nextMode);
    });

    window.addEventListener("storage", (event) => {
      if (event.key === storageKey && isThemeMode(event.newValue)) {
        applyMode(event.newValue, toggle);
      }
    });
  });
})();
