import { useMemo } from "react";
import { CssBaseline, ThemeProvider } from "@mui/material";
import { RouterProvider, createHashRouter } from "react-router-dom";

import { buildTheme } from "@/theme";
import { useSettingsStore } from "@/stores/settingsStore";
import LoginPage from "@/pages/Login/LoginPage";
import CreateProfilePage from "@/pages/CreateProfile/CreateProfilePage";
import FilesPage from "@/pages/Files/FilesPage";
import SettingsPage from "@/pages/Settings/SettingsPage";
import Gate from "@/components/Gate";
import SnackbarHost from "@/components/SnackbarHost";
import AppEvents from "@/components/AppEvents";
import HostKeyDialog from "@/components/HostKeyDialog";
import DialogHost from "@/components/DialogHost";

const router = createHashRouter([
  { path: "/", element: <Gate /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/create-profile", element: <CreateProfilePage /> },
  { path: "/files/*", element: <FilesPage /> },
  { path: "/settings", element: <SettingsPage /> },
]);

export default function App() {
  const themeMode = useSettingsStore((s) => s.themeMode);
  const theme = useMemo(() => buildTheme(themeMode), [themeMode]);
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppEvents />
      <RouterProvider router={router} />
      <SnackbarHost />
      <HostKeyDialog />
      <DialogHost />
    </ThemeProvider>
  );
}
