import { Snackbar, Alert } from "@mui/material";
import { useSnackbarStore } from "@/stores/snackbarStore";

export default function SnackbarHost() {
  const current = useSnackbarStore((s) => s.current);
  const dismiss = useSnackbarStore((s) => s.dismiss);
  return (
    <Snackbar
      open={Boolean(current)}
      autoHideDuration={current?.duration ?? 3000}
      onClose={(_, reason) => {
        if (reason === "clickaway") return;
        dismiss();
      }}
      anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
    >
      {current ? (
        <Alert
          severity={current.severity}
          variant="filled"
          onClose={dismiss}
          sx={{ minWidth: 280 }}
        >
          {current.text}
        </Alert>
      ) : undefined}
    </Snackbar>
  );
}
