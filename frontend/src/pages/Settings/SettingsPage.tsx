import { Box, Button, Stack, ToggleButton, ToggleButtonGroup, Typography } from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import LightModeIcon from "@mui/icons-material/LightMode";
import DarkModeIcon from "@mui/icons-material/DarkMode";
import { useNavigate } from "react-router-dom";

import { useSettingsStore } from "@/stores/settingsStore";

export default function SettingsPage() {
  const navigate = useNavigate();
  const mode = useSettingsStore((s) => s.themeMode);
  const setMode = useSettingsStore((s) => s.setThemeMode);

  return (
    <Box sx={{ p: 4, maxWidth: 720 }}>
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 3 }}>
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(-1)}>
          Back
        </Button>
        <Typography variant="h5" fontWeight={600}>
          Settings
        </Typography>
      </Stack>

      <Stack spacing={3}>
        <Box>
          <Typography variant="subtitle1" gutterBottom>
            Theme
          </Typography>
          <ToggleButtonGroup
            value={mode}
            exclusive
            onChange={(_, next) => next && setMode(next)}
            color="primary"
            size="small"
          >
            <ToggleButton value="light">
              <LightModeIcon fontSize="small" sx={{ mr: 1 }} /> Light
            </ToggleButton>
            <ToggleButton value="dark">
              <DarkModeIcon fontSize="small" sx={{ mr: 1 }} /> Dark
            </ToggleButton>
          </ToggleButtonGroup>
        </Box>

        <Box>
          <Typography variant="subtitle1" gutterBottom>
            About
          </Typography>
          <Typography variant="body2" color="text.secondary">
            PIdisk runs locally. Profiles are stored in a bbolt file in your
            user data directory; passphrases are kept in the OS keyring under
            the service "com.pidisk.profiles".
          </Typography>
        </Box>
      </Stack>
    </Box>
  );
}
