import { createTheme, ThemeOptions, alpha } from "@mui/material";

export type ThemeMode = "light" | "dark";

const sharedTypography: ThemeOptions["typography"] = {
  fontFamily:
    '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  fontSize: 13,
};

const sharedShape: ThemeOptions["shape"] = { borderRadius: 10 };

export const buildTheme = (mode: ThemeMode) => {
  const isDark = mode === "dark";

  const palette = {
    mode,
    primary: { main: isDark ? "#7aa7ff" : "#1c74d4" },
    secondary: { main: isDark ? "#c084fc" : "#7c3aed" },
    background: isDark
      ? { default: "#11141b", paper: "#1a1f2b" }
      : { default: "#f6f7fa", paper: "#ffffff" },
    text: isDark
      ? { primary: "#e7e9ee", secondary: "#a4abbb", disabled: "#6b7280" }
      : { primary: "#1d2230", secondary: "#4b5263", disabled: "#9aa3b2" },
    divider: isDark ? "rgba(255,255,255,0.14)" : "rgba(0,0,0,0.12)",
  } as const;

  const borderColor = isDark
    ? "rgba(255,255,255,0.28)"
    : "rgba(0,0,0,0.23)";
  const borderHover = isDark
    ? "rgba(255,255,255,0.55)"
    : "rgba(0,0,0,0.55)";
  const inputBg = isDark ? "rgba(255,255,255,0.04)" : "rgba(0,0,0,0.02)";

  return createTheme({
    palette,
    typography: sharedTypography,
    shape: sharedShape,
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            backgroundColor: palette.background.default,
            color: palette.text.primary,
          },
        },
      },
      MuiButton: { defaultProps: { disableElevation: true } },
      MuiPaper: {
        defaultProps: { elevation: 0, variant: "outlined" },
        styleOverrides: {
          root: {
            backgroundImage: "none",
            backgroundColor: palette.background.paper,
            borderColor: palette.divider,
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            backgroundColor: palette.background.paper,
            borderBottom: `1px solid ${palette.divider}`,
            color: palette.text.primary,
          },
        },
      },
      MuiToolbar: {
        styleOverrides: {
          root: { backgroundColor: "transparent" },
        },
      },
      MuiOutlinedInput: {
        styleOverrides: {
          root: {
            backgroundColor: inputBg,
            "& .MuiOutlinedInput-notchedOutline": {
              borderColor,
            },
            "&:hover .MuiOutlinedInput-notchedOutline": {
              borderColor: borderHover,
            },
          },
          input: { color: palette.text.primary },
        },
      },
      MuiInputLabel: {
        styleOverrides: {
          root: {
            color: palette.text.secondary,
            "&.Mui-focused": { color: palette.primary.main },
          },
        },
      },
      MuiFormHelperText: {
        styleOverrides: { root: { color: palette.text.secondary } },
      },
      MuiIconButton: {
        styleOverrides: {
          root: { color: palette.text.primary },
        },
      },
      MuiDivider: {
        styleOverrides: { root: { borderColor: palette.divider } },
      },
      MuiTooltip: {
        styleOverrides: {
          tooltip: {
            backgroundColor: isDark ? "#2a3142" : "#1d2230",
            color: "#ffffff",
            fontSize: 12,
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            backgroundColor: palette.background.paper,
            borderColor: palette.divider,
          },
        },
      },
      MuiDialog: {
        styleOverrides: {
          paper: { backgroundColor: palette.background.paper },
        },
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            backgroundColor: palette.background.paper,
            borderColor: palette.divider,
          },
        },
      },
      MuiListItem: {
        styleOverrides: {
          root: {
            "&:hover": {
              backgroundColor: alpha(palette.primary.main, isDark ? 0.12 : 0.06),
            },
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            backgroundColor: isDark ? "rgba(255,255,255,0.08)" : "rgba(0,0,0,0.06)",
          },
        },
      },
    },
  });
};
