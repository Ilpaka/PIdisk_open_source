import { useEffect, useState, useCallback } from "react";
import { Box, IconButton, Stack, Typography } from "@mui/material";
import FolderIcon from "@mui/icons-material/Folder";
import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import ChevronRightIcon from "@mui/icons-material/ChevronRight";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";

import { filesApi } from "@/api/files";
import type { FileEntry } from "@/types/domain";

interface TreeNode {
  path: string;
  name: string;
  loaded: boolean;
  expanded: boolean;
  children: TreeNode[];
}

interface Props {
  rootPath: string;
  currentPath: string;
  onNavigate: (path: string) => void;
  onMove?: (src: string, dst: string) => Promise<void>;
}

function makeNode(path: string, name: string): TreeNode {
  return { path, name, loaded: false, expanded: false, children: [] };
}

function updateNode(
  root: TreeNode,
  target: string,
  fn: (n: TreeNode) => TreeNode,
): TreeNode {
  if (root.path === target) return fn(root);
  return {
    ...root,
    children: root.children.map((c) => updateNode(c, target, fn)),
  };
}

export default function FolderTree({ rootPath, currentPath, onNavigate, onMove }: Props) {
  const [root, setRoot] = useState<TreeNode>(makeNode(rootPath, rootPath));

  const load = useCallback(async (node: TreeNode): Promise<TreeNode> => {
    if (node.loaded) return node;
    try {
      const listing = await filesApi.readDir(node.path);
      const children = (listing.entries ?? [])
        .filter((e: FileEntry) => e.isDir)
        .map((e) => makeNode(e.path, e.name));
      return { ...node, loaded: true, children };
    } catch {
      return { ...node, loaded: true, children: [] };
    }
  }, []);

  useEffect(() => {
    setRoot(makeNode(rootPath, rootPath));
  }, [rootPath]);

  useEffect(() => {
    let alive = true;
    (async () => {
      const loaded = await load({ ...root, expanded: true });
      if (alive) setRoot({ ...loaded, expanded: true });
    })();
    return () => {
      alive = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rootPath]);

  const toggle = async (node: TreeNode) => {
    let nextNode = node;
    if (!node.loaded) {
      nextNode = await load(node);
    }
    setRoot((prev) =>
      updateNode(prev, node.path, () => ({
        ...nextNode,
        expanded: !node.expanded,
      })),
    );
  };

  const handleDrop = async (target: string, ev: React.DragEvent) => {
    if (!onMove) return;
    ev.preventDefault();
    const src = ev.dataTransfer.getData("application/x-pidisk-file");
    if (!src) return;
    const fileName = src.split("/").pop();
    if (!fileName) return;
    await onMove(src, `${target}/${fileName}`);
  };

  const renderNode = (node: TreeNode, depth: number) => {
    const isCurrent = node.path === currentPath;
    return (
      <Box key={node.path}>
        <Stack
          direction="row"
          alignItems="center"
          spacing={0.5}
          sx={{
            pl: depth * 1.5,
            py: 0.25,
            cursor: "pointer",
            borderRadius: 1,
            bgcolor: isCurrent ? "action.selected" : "transparent",
            "&:hover": { bgcolor: "action.hover" },
          }}
          onClick={() => onNavigate(node.path)}
          onDragOver={(e) => {
            if (onMove) e.preventDefault();
          }}
          onDrop={(e) => handleDrop(node.path, e)}
        >
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              void toggle(node);
            }}
            sx={{ p: 0.25 }}
          >
            {node.expanded ? (
              <KeyboardArrowDownIcon fontSize="inherit" />
            ) : (
              <ChevronRightIcon fontSize="inherit" />
            )}
          </IconButton>
          {node.expanded ? (
            <FolderOpenIcon fontSize="small" color="primary" />
          ) : (
            <FolderIcon fontSize="small" color="primary" />
          )}
          <Typography variant="body2" noWrap>
            {node.name === rootPath ? rootPath : node.name}
          </Typography>
        </Stack>
        {node.expanded && node.children.length > 0
          ? node.children.map((c) => renderNode(c, depth + 1))
          : null}
      </Box>
    );
  };

  return <Box sx={{ p: 1, overflow: "auto" }}>{renderNode(root, 0)}</Box>;
}
