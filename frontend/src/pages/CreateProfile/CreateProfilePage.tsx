import { FormEvent, useEffect, useState } from "react";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Container,
  Divider,
  Grid,
  IconButton,
  InputAdornment,
  Paper,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import VisibilityIcon from "@mui/icons-material/Visibility";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import { useNavigate } from "react-router-dom";

import type { ProfileInput } from "@/types/domain";
import { useProfileStore } from "@/stores/profileStore";
import { useSnackbarStore } from "@/stores/snackbarStore";
import { profilesApi, type ProfileDefaults } from "@/api/profiles";
import GeneratedKeyDialog from "@/pages/CreateProfile/GeneratedKeyDialog";

// WebKit on macOS auto-capitalizes the first character and auto-corrects in
// regular text inputs. Apply this to identifier-like fields (host, username,
// paths) so "root" doesn't silently become "Root".
const identifierInputProps = {
  autoCapitalize: "none",
  autoCorrect: "off",
  spellCheck: false,
} as const;

const initial: ProfileInput = {
  name: "",
  host: "",
  port: 22,
  username: "",
  privateKeyPath: "",
  passphrase: "",
  rootDir: "",
  trashDir: "",
  localSyncDir: "",
};

export default function CreateProfilePage() {
  const navigate = useNavigate();
  const [values, setValues] = useState<ProfileInput>(initial);
  const [defaults, setDefaults] = useState<ProfileDefaults | null>(null);
  const [showPass, setShowPass] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [generated, setGenerated] = useState<{
    publicKey: string;
    keyPath: string;
    username: string;
    host: string;
  } | null>(null);

  const create = useProfileStore((s) => s.create);
  const pushSnack = useSnackbarStore((s) => s.push);

  useEffect(() => {
    let cancelled = false;
    const t = setTimeout(async () => {
      try {
        const d = await profilesApi.suggestDefaults(values.name, values.username);
        if (!cancelled) setDefaults(d);
      } catch {
        // ignore
      }
    }, 150);
    return () => {
      cancelled = true;
      clearTimeout(t);
    };
  }, [values.name, values.username]);

  const update =
    (field: keyof ProfileInput) =>
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = field === "port" ? Number(e.target.value || 0) : e.target.value;
      setValues((prev) => ({ ...prev, [field]: value }));
    };

  const customKey = values.privateKeyPath.trim() !== "";
  const effectiveRoot = values.rootDir || defaults?.remoteRoot || "";
  const effectiveTrash = values.trashDir || defaults?.remoteTrash || "";
  const effectiveSync = values.localSyncDir || defaults?.localSyncDir || "";
  const keyHint = customKey
    ? values.privateKeyPath
    : "A fresh Ed25519 key will be generated in ~/.ssh.";

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      const result = await create({ ...values, port: Number(values.port) || 22 });
      pushSnack("success", "Profile created");
      if (result.generatedPublicKey) {
        setGenerated({
          publicKey: result.generatedPublicKey,
          keyPath: result.generatedKeyPath ?? "",
          username: values.username,
          host: values.host,
        });
      } else {
        navigate("/login");
      }
    } catch (err) {
      pushSnack("error", String(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Container maxWidth="md" sx={{ py: 5 }}>
      <Paper variant="outlined" sx={{ p: 4 }}>
        <Stack spacing={3}>
          <Stack direction="row" alignItems="center" spacing={2}>
            <IconButton onClick={() => navigate("/login")} aria-label="back to login">
              <ArrowBackIcon />
            </IconButton>
            <Typography variant="h5" fontWeight={600}>
              New SSH profile
            </Typography>
          </Stack>
          <Typography variant="body2" color="text.secondary">
            Fill in the connection details. PIdisk auto-generates an SSH key,
            picks the remote home directory, the trash folder and the local
            sync folder. Override anything under "Advanced" if you need to.
          </Typography>
          <Divider />

          <Box component="form" onSubmit={onSubmit}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  required
                  label="Profile name"
                  value={values.name}
                  onChange={update("name")}
                  helperText="Label shown in the profile list."
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  required
                  label="Username"
                  value={values.username}
                  onChange={update("username")}
                  placeholder="root"
                  inputProps={identifierInputProps}
                />
              </Grid>
              <Grid item xs={8} md={9}>
                <TextField
                  fullWidth
                  required
                  label="Host"
                  value={values.host}
                  onChange={update("host")}
                  placeholder="example.com or 1.2.3.4"
                  inputProps={identifierInputProps}
                />
              </Grid>
              <Grid item xs={4} md={3}>
                <TextField
                  fullWidth
                  required
                  label="Port"
                  type="number"
                  value={values.port}
                  onChange={update("port")}
                  inputProps={{ min: 1, max: 65535, ...identifierInputProps }}
                />
              </Grid>

              <Grid item xs={12}>
                <Accordion
                  expanded={advancedOpen}
                  onChange={(_, v) => setAdvancedOpen(v)}
                  disableGutters
                  square
                  sx={{ bgcolor: "transparent" }}
                >
                  <AccordionSummary expandIcon={<ExpandMoreIcon />} sx={{ px: 0 }}>
                    <Stack>
                      <Typography variant="subtitle2">Advanced</Typography>
                      <Typography variant="caption" color="text.secondary">
                        Key: {keyHint}
                        {"  ·  "}Remote: {effectiveRoot}
                        {"  ·  "}Trash: {effectiveTrash}
                        {"  ·  "}Local sync: {effectiveSync || "(skipped)"}
                      </Typography>
                    </Stack>
                  </AccordionSummary>
                  <AccordionDetails sx={{ px: 0 }}>
                    <Stack spacing={2}>
                      <Alert severity="info">
                        Leave "Private key path" empty to have PIdisk generate
                        a fresh Ed25519 key and store its random passphrase in
                        the OS keyring. Provide a path here only if you want
                        to reuse an existing key.
                      </Alert>
                      <TextField
                        fullWidth
                        label="Private key path (optional)"
                        value={values.privateKeyPath}
                        onChange={update("privateKeyPath")}
                        placeholder="/Users/you/.ssh/id_ed25519"
                        helperText="Absolute path. Leave empty to auto-generate."
                        inputProps={identifierInputProps}
                      />
                      <TextField
                        fullWidth
                        label="Existing key passphrase (optional)"
                        type={showPass ? "text" : "password"}
                        value={values.passphrase}
                        onChange={update("passphrase")}
                        helperText="Only needed if the private key path above points at an encrypted key."
                        inputProps={identifierInputProps}
                        InputProps={{
                          endAdornment: (
                            <InputAdornment position="end">
                              <Tooltip title={showPass ? "Hide" : "Show"}>
                                <IconButton onClick={() => setShowPass((v) => !v)} edge="end">
                                  {showPass ? <VisibilityOffIcon /> : <VisibilityIcon />}
                                </IconButton>
                              </Tooltip>
                            </InputAdornment>
                          ),
                        }}
                      />
                      <TextField
                        fullWidth
                        label="Remote root"
                        value={values.rootDir}
                        onChange={update("rootDir")}
                        placeholder={defaults?.remoteRoot || "/home/user"}
                        helperText="Default starting directory on the server."
                        inputProps={identifierInputProps}
                      />
                      <TextField
                        fullWidth
                        label="Remote trash"
                        value={values.trashDir}
                        onChange={update("trashDir")}
                        placeholder={defaults?.remoteTrash || "/home/user/.pidisk-trash"}
                        helperText="Where deletions are moved. Created on first delete."
                        inputProps={identifierInputProps}
                      />
                      <TextField
                        fullWidth
                        label="Local sync directory"
                        value={values.localSyncDir}
                        onChange={update("localSyncDir")}
                        placeholder={defaults?.localSyncDir || "~/PIdiskSync/<name>"}
                        helperText="Folder on this machine used by the sync engine. Created automatically."
                        inputProps={identifierInputProps}
                      />
                    </Stack>
                  </AccordionDetails>
                </Accordion>
              </Grid>

              <Grid item xs={12}>
                <Stack direction="row" spacing={2} justifyContent="flex-end">
                  <Button variant="outlined" onClick={() => navigate("/login")}>
                    Cancel
                  </Button>
                  <Button type="submit" variant="contained" disabled={submitting}>
                    {submitting ? "Saving..." : "Create profile"}
                  </Button>
                </Stack>
              </Grid>
            </Grid>
          </Box>
        </Stack>
      </Paper>
      <GeneratedKeyDialog
        open={Boolean(generated)}
        publicKey={generated?.publicKey ?? ""}
        keyPath={generated?.keyPath ?? ""}
        username={generated?.username ?? ""}
        host={generated?.host ?? ""}
        onClose={() => {
          setGenerated(null);
          navigate("/login");
        }}
      />
    </Container>
  );
}
