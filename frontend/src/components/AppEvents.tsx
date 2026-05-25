import { useEffect } from "react";
import { on } from "@/api/events";
import { useConnectionStore } from "@/stores/connectionStore";
import { useTransferStore } from "@/stores/transferStore";
import { useSyncStore } from "@/stores/syncStore";

export default function AppEvents() {
  const setState = useConnectionStore((s) => s.setState);
  const showPrompt = useConnectionStore((s) => s.showHostKeyPrompt);
  const upsert = useTransferStore((s) => s.upsert);
  const complete = useTransferStore((s) => s.complete);
  const setStats = useSyncStore((s) => s.setStats);

  useEffect(() => {
    const offs = [
      on("connection:state", setState),
      on("hostkey:prompt", showPrompt),
      on("transfer:started", upsert),
      on("transfer:progress", upsert),
      on("transfer:done", complete),
      on("sync:status", setStats),
    ];
    return () => offs.forEach((off) => off());
  }, [setState, showPrompt, upsert, complete, setStats]);

  return null;
}
