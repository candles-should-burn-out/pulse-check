import { createContext, useContext } from "react";

import { type AppThemeMode } from "./theme";

export type AppThemeModeContextValue = {
  mode: AppThemeMode;
  toggleMode: () => void;
};

export const AppThemeModeContext =
  createContext<AppThemeModeContextValue | null>(null);

export function useThemeMode() {
  const value = useContext(AppThemeModeContext);

  if (!value) {
    throw new Error("useThemeMode must be used inside AppThemeModeProvider");
  }

  return value;
}
