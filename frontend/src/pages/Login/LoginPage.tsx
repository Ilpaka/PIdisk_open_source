import { useEffect, useState } from "react";
import {
  Avatar,
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Container,
  Divider,
  IconButton,
  Stack,
  Typography,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import StorageIcon from "@mui/icons-material/Storage";
import { useNavigate } from "react-router-dom";

import type { Profile } from "@/types/domain";
import { useProfileStore } from "@/stores/profileStore";
import { useSnackbarStore } from "@/stores/snackbarStore";
import { connectionApi } from "@/api/connection";
import { profilesApi } from "@/api/profiles";
import { confirmDialog } from "@/stores/dialogStore";
import PassphrasePrompt from "@/pages/Login/PassphrasePrompt";

export default function LoginPage() {
  const navigate = useNavigate();
  const { profiles, status, refresh, remove, markActive } = useProfileStore();
  const pushSnack = useSnackbarStore((s) => s.push);
  const [pending, setPending] = useState<Profile | null>(null);

  useEffect(() => {
    if (status === "loading") void refresh();
  }, [status, refresh]);

  const attemptConnect = async (profile: Profile, passphrase: string) => {
    const result = await connectionApi.unlock(profile.id, passphrase);
    if (!result.connected) {
      throw new Error("Connection refused by server");
    }
    // The connection usecase already marks the profile active on the Go side;
    // mirror it into the frontend store so FilesPage does not bounce back to /login.
    markActive(result.profile);
    pushSnack("success", `Connected as ${profile.username}@${profile.host}`);
    navigate("/files");
  };

  const onSelect = async (profile: Profile) => {
    const cached = await profilesApi.hasPassphrase(profile.id);
    if (cached) {
      try {
        await attemptConnect(profile, "");
      } catch (err) {
        pushSnack("warning", `Stored passphrase rejected: ${String(err)}`);
        setPending(profile);
      }
      return;
    }
    setPending(profile);
  };

  const onDelete = async (id: string, name: string) => {
    const ok = await confirmDialog({
      title: "Delete profile",
      message: `Delete profile "${name}"? The stored passphrase and the auto-generated SSH key file will be removed.`,
      confirmText: "Delete",
      destructive: true,
    });
    if (!ok) return;
    try {
      await remove(id);
      pushSnack("success", `Profile "${name}" deleted`);
    } catch (err) {
      pushSnack("error", `Cannot delete profile: ${String(err)}`);
    }
  };

  return (
    <Container maxWidth="sm" sx={{ py: 8 }}>
      <Stack spacing={4}>
        <Stack alignItems="center" spacing={1}>
          <Typography variant="h4" fontWeight={600}>
            PIdisk
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Pick a profile to connect, or add a new one.
          </Typography>
        </Stack>

        {profiles.length === 0 ? (
          <Stack spacing={2} alignItems="center">
            <Typography color="text.secondary">No profiles yet.</Typography>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => navigate("/create-profile")}
            >
              Create profile
            </Button>
          </Stack>
        ) : (
          <Stack spacing={1.5}>
            {profiles.map((p) => (
              <Card key={p.id} variant="outlined">
                <Stack direction="row" alignItems="stretch">
                  <CardActionArea onClick={() => onSelect(p)} sx={{ flex: 1 }}>
                    <CardContent>
                      <Stack direction="row" spacing={2} alignItems="center">
                        <Avatar variant="rounded" sx={{ bgcolor: "primary.main" }}>
                          <StorageIcon />
                        </Avatar>
                        <Box>
                          <Typography fontWeight={600}>{p.name}</Typography>
                          <Typography variant="caption" color="text.secondary">
                            {p.username}@{p.host}:{p.port} ({p.rootDir})
                          </Typography>
                        </Box>
                      </Stack>
                    </CardContent>
                  </CardActionArea>
                  <Divider orientation="vertical" flexItem />
                  <Box sx={{ display: "flex", alignItems: "center", px: 1 }}>
                    <IconButton
                      onClick={() => onDelete(p.id, p.name)}
                      aria-label={`delete profile ${p.name}`}
                    >
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
                  </Box>
                </Stack>
              </Card>
            ))}
            <Button
              variant="text"
              startIcon={<AddIcon />}
              onClick={() => navigate("/create-profile")}
            >
              Add another profile
            </Button>
          </Stack>
        )}
      </Stack>
      <PassphrasePrompt
        profile={pending}
        onCancel={() => setPending(null)}
        onSubmit={async (p) => {
          if (!pending) return;
          await attemptConnect(pending, p);
          setPending(null);
        }}
      />
    </Container>
  );
}
