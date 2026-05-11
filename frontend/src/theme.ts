import { createTheme } from "@mui/material/styles";

export const appTheme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: "#235c63",
      contrastText: "#ffffff",
    },
    secondary: {
      main: "#8a4f3d",
    },
    background: {
      default: "#f6f7f4",
      paper: "#ffffff",
    },
    success: {
      main: "#39705f",
    },
    warning: {
      main: "#b7791f",
    },
    error: {
      main: "#b42318",
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
