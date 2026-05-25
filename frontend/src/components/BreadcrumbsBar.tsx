import { Breadcrumbs, IconButton, Link, Stack, Tooltip } from "@mui/material";
import HomeIcon from "@mui/icons-material/Home";
import RefreshIcon from "@mui/icons-material/Refresh";

interface Props {
  cwd: string;
  rootDir: string;
  onNavigate: (path: string) => void;
  onRefresh: () => void;
}

export default function BreadcrumbsBar({ cwd, rootDir, onNavigate, onRefresh }: Props) {
  const segments = cwd && rootDir && cwd.startsWith(rootDir)
    ? cwd.slice(rootDir.length).split("/").filter(Boolean)
    : cwd.split("/").filter(Boolean);

  const crumbs: { label: string; path: string }[] = [];
  let acc = rootDir || "/";
  for (const seg of segments) {
    acc = acc === "/" ? `/${seg}` : `${acc}/${seg}`;
    crumbs.push({ label: seg, path: acc });
  }

  return (
    <Stack direction="row" spacing={1} alignItems="center" sx={{ px: 2, py: 1 }}>
      <Tooltip title="Refresh">
        <IconButton onClick={onRefresh} size="small">
          <RefreshIcon fontSize="small" />
        </IconButton>
      </Tooltip>
      <Breadcrumbs separator="/" aria-label="path">
        <Link
          component="button"
          color="inherit"
          underline="hover"
          onClick={() => onNavigate(rootDir || "/")}
          sx={{ display: "flex", alignItems: "center", gap: 0.5 }}
        >
          <HomeIcon fontSize="inherit" />
          {rootDir || "/"}
        </Link>
        {crumbs.map((c) => (
          <Link
            key={c.path}
            component="button"
            color="inherit"
            underline="hover"
            onClick={() => onNavigate(c.path)}
          >
            {c.label}
          </Link>
        ))}
      </Breadcrumbs>
    </Stack>
  );
}
