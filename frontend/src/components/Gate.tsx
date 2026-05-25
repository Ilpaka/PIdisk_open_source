import { useEffect } from "react";
import { Box, CircularProgress } from "@mui/material";
import { Navigate } from "react-router-dom";
import { useProfileStore } from "@/stores/profileStore";

/**
 * Gate decides where to land based on the profile store status.
 * Loading + initial load happens here so the rest of the UI can assume a
 * resolved state.
 */
export default function Gate() {
  const status = useProfileStore((s) => s.status);
  const refresh = useProfileStore((s) => s.refresh);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  if (status === "loading") {
    return (
      <Box sx={{ display: "grid", placeItems: "center", height: "100vh" }}>
        <CircularProgress />
      </Box>
    );
  }
  if (status === "none") return <Navigate to="/create-profile" replace />;
  if (status === "active") return <Navigate to="/files" replace />;
  return <Navigate to="/login" replace />;
}
