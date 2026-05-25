import { ReactNode, useRef, useState } from "react";
import { Box } from "@mui/material";

interface Props {
  children: ReactNode;
  storageKey?: string;
  defaultWidth?: number;
  minWidth?: number;
  maxWidth?: number;
}

/**
 * Two-column layout with a draggable vertical splitter on the right edge of
 * the sidebar. The chosen width persists to localStorage so the layout is
 * restored across sessions.
 */
export default function ResizableSidebar({
  children,
  storageKey = "pidisk.sidebar.width",
  defaultWidth = 260,
  minWidth = 180,
  maxWidth = 640,
}: Props) {
  const initial = (() => {
    if (typeof window === "undefined") return defaultWidth;
    const raw = Number(window.localStorage?.getItem(storageKey));
    if (Number.isFinite(raw) && raw >= minWidth && raw <= maxWidth) return raw;
    return defaultWidth;
  })();

  const widthRef = useRef(initial);
  const [width, setWidth] = useState(initial);

  const onMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    const startX = e.clientX;
    const startWidth = widthRef.current;

    const onMove = (ev: MouseEvent) => {
      const next = Math.max(
        minWidth,
        Math.min(maxWidth, startWidth + (ev.clientX - startX)),
      );
      widthRef.current = next;
      setWidth(next);
    };
    const onUp = () => {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      try {
        window.localStorage.setItem(storageKey, String(widthRef.current));
      } catch {
        // ignore persistence failure
      }
    };

    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  };

  const onDoubleClick = () => {
    widthRef.current = defaultWidth;
    setWidth(defaultWidth);
    try {
      window.localStorage.setItem(storageKey, String(defaultWidth));
    } catch {
      // ignore
    }
  };

  return (
    <Box sx={{ display: "flex", flexShrink: 0, height: "100%" }}>
      <Box
        sx={{
          width,
          minWidth,
          maxWidth,
          overflow: "auto",
          height: "100%",
        }}
      >
        {children}
      </Box>
      <Box
        role="separator"
        aria-orientation="vertical"
        aria-label="Resize sidebar"
        onMouseDown={onMouseDown}
        onDoubleClick={onDoubleClick}
        sx={{
          width: 6,
          flexShrink: 0,
          cursor: "col-resize",
          position: "relative",
          bgcolor: "divider",
          "&:hover": { bgcolor: "primary.main" },
          "&:active": { bgcolor: "primary.main" },
          transition: "background-color 0.15s ease",
        }}
      />
    </Box>
  );
}
