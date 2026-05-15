import {
  PropsWithChildren,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import Keycloak, { KeycloakInstance } from "keycloak-js";

import { AuthContext } from "./context";
import type { AuthContextValue, AuthStatus } from "./context";

const keycloakConfig = {
  url: import.meta.env.VITE_KEYCLOAK_URL ?? "http://localhost:8081",
  realm: import.meta.env.VITE_KEYCLOAK_REALM ?? "pulse-check",
  clientId:
    import.meta.env.VITE_KEYCLOAK_CLIENT_ID ?? "pulse-check-frontend",
};

const keycloak: KeycloakInstance = new Keycloak(keycloakConfig);
const appPublicUrl =
  import.meta.env.VITE_APP_PUBLIC_URL?.replace(/\/$/, "") ??
  window.location.origin;

let initPromise: Promise<boolean> | null = null;

export function AuthProvider({ children }: PropsWithChildren) {
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [userName, setUserName] = useState<string | null>(null);

  useEffect(() => {
    if (!initPromise) {
      initPromise = keycloak.init({
        pkceMethod: "S256",
        responseMode: "query",
        checkLoginIframe: false,
      });
    }

    initPromise
      .then((authenticated) => {
        setStatus(authenticated ? "authenticated" : "anonymous");
        setUserName(readUserName(keycloak));
      })
      .catch((error) => {
        console.error("Keycloak init failed", error);
        setStatus("error");
        setUserName(null);
      });
  }, []);

  const login = useCallback(async () => {
    await keycloak.login({
      redirectUri: `${appPublicUrl}/app/`,
    });
  }, []);

  const logout = useCallback(async () => {
    await keycloak.logout({
      redirectUri: appPublicUrl,
    });
  }, []);

  const getAccessToken = useCallback(async () => {
    if (!keycloak.authenticated) {
      await keycloak.login({
        redirectUri: `${appPublicUrl}/app/`,
      });
      throw new Error("Пользователь не авторизован");
    }

    await keycloak.updateToken(30);

    if (!keycloak.token) {
      throw new Error("Не удалось получить access token");
    }

    return keycloak.token;
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      status,
      userName,
      login,
      logout,
      getAccessToken,
    }),
    [getAccessToken, login, logout, status, userName]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

function readUserName(instance: KeycloakInstance): string | null {
  const tokenParsed = instance.tokenParsed;
  if (!tokenParsed) {
    return null;
  }

  const preferredUsername = tokenParsed.preferred_username;
  if (typeof preferredUsername === "string" && preferredUsername !== "") {
    return preferredUsername;
  }

  const email = tokenParsed.email;
  return typeof email === "string" && email !== "" ? email : null;
}
