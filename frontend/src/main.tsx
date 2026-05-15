import "./telemetry";

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import App from "./App";
import { AuthProvider } from "./auth/keycloak";
import { AppThemeModeProvider } from "./theme-mode";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppThemeModeProvider>
      <BrowserRouter basename="/app">
        <AuthProvider>
          <App />
        </AuthProvider>
      </BrowserRouter>
    </AppThemeModeProvider>
  </StrictMode>
);
