import { type ReactNode, useEffect, useMemo, useState } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";

import {
  AppThemeModeContext,
  type AppThemeModeContextValue,
} from "./theme-mode-context";
import {
  APP_THEME_STORAGE_KEY,
  type AppThemeMode,
  applyDocumentThemeMode,
  createAppTheme,
  getStoredThemeMode,
  isAppThemeMode,
  storeThemeMode,
} from "./theme";

export function AppThemeModeProvider({ children }: { children: ReactNode }) {
  const [mode, setMode] = useState<AppThemeMode>(() => getStoredThemeMode());
  const theme = useMemo(() => createAppTheme(mode), [mode]);

  useEffect(() => {
    applyDocumentThemeMode(mode);
    storeThemeMode(mode);
  }, [mode]);

  useEffect(() => {
    const handleStorage = (event: StorageEvent) => {
      if (
        event.key === APP_THEME_STORAGE_KEY &&
        isAppThemeMode(event.newValue)
      ) {
        setMode(event.newValue);
      }
    };

    window.addEventListener("storage", handleStorage);

    return () => {
      window.removeEventListener("storage", handleStorage);
    };
  }, []);

  const value = useMemo<AppThemeModeContextValue>(
    () => ({
      mode,
      toggleMode: () => {
        setMode((currentMode) => (currentMode === "dark" ? "light" : "dark"));
      },
    }),
    [mode]
  );

  return (
    <AppThemeModeContext.Provider value={value}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </AppThemeModeContext.Provider>
  );
}
