import { createContext } from "react";

export type AuthStatus = "loading" | "authenticated" | "anonymous" | "error";

export type AuthContextValue = {
  status: AuthStatus;
  userName: string | null;
  login: () => Promise<void>;
  logout: () => Promise<void>;
  getAccessToken: () => Promise<string>;
};

export const AuthContext = createContext<AuthContextValue | null>(null);
