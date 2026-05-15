import { createTheme } from "@mui/material/styles";

export type AppThemeMode = "light" | "dark";

export const APP_THEME_STORAGE_KEY = "pulse-check-theme-mode";

export function isAppThemeMode(value: string | null): value is AppThemeMode {
  return value === "light" || value === "dark";
}

export function getStoredThemeMode(): AppThemeMode {
  if (typeof window === "undefined") {
    return "light";
  }

  const cookieMode = readThemeModeCookie();

  if (cookieMode) {
    return cookieMode;
  }

  try {
    const storedMode = window.localStorage.getItem(APP_THEME_STORAGE_KEY);
    return isAppThemeMode(storedMode) ? storedMode : "light";
  } catch {
    return "light";
  }
}

export function applyDocumentThemeMode(mode: AppThemeMode) {
  if (typeof document === "undefined") {
    return;
  }

  document.documentElement.dataset.theme = mode;
  document.documentElement.style.colorScheme = mode;
}

export function storeThemeMode(mode: AppThemeMode) {
  try {
    window.localStorage.setItem(APP_THEME_STORAGE_KEY, mode);
  } catch {
    // Theme selection is still applied for the current page.
  }

  writeThemeModeCookie(mode);
}

function readThemeModeCookie(): AppThemeMode | null {
  if (typeof document === "undefined") {
    return null;
  }

  const cookie = document.cookie
    .split("; ")
    .find((entry) => entry.startsWith(`${APP_THEME_STORAGE_KEY}=`));
  const mode = cookie?.split("=")[1] ?? null;

  return isAppThemeMode(mode) ? mode : null;
}

function writeThemeModeCookie(mode: AppThemeMode) {
  if (typeof document === "undefined") {
    return;
  }

  document.cookie = `${APP_THEME_STORAGE_KEY}=${mode}; path=/; max-age=31536000; SameSite=Lax`;
}

export function createAppTheme(mode: AppThemeMode) {
  const isDark = mode === "dark";

  return createTheme({
    palette: {
      mode,
      primary: {
        main: isDark ? "#88c0d0" : "#5e81ac",
        contrastText: isDark ? "#2e3440" : "#eceff4",
      },
      secondary: {
        main: "#b48ead",
      },
      background: {
        default: isDark ? "#2e3440" : "#eceff4",
        paper: isDark ? "#3b4252" : "#ffffff",
      },
      success: {
        main: "#a3be8c",
      },
      warning: {
        main: "#ebcb8b",
      },
      error: {
        main: "#bf616a",
      },
      divider: isDark ? "rgba(216, 222, 233, 0.18)" : "#d8dee9",
      text: {
        primary: isDark ? "#eceff4" : "#2e3440",
        secondary: isDark ? "#d8dee9" : "#4c566a",
      },
    },
    shape: {
      borderRadius: 8,
    },
    typography: {
      fontFamily:
        'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
      h1: {
        fontSize: "2rem",
        fontWeight: 700,
        letterSpacing: 0,
      },
      h2: {
        fontSize: "1.35rem",
        fontWeight: 700,
        letterSpacing: 0,
      },
      button: {
        fontWeight: 700,
        textTransform: "none",
        letterSpacing: 0,
      },
    },
    components: {
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            minHeight: 40,
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
          },
        },
      },
    },
  });
}
