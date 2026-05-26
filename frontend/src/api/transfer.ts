import {
  CancelTransfer,
  DownloadFile,
  ListTransfers,
  UploadFile,
} from "../../wailsjs/go/wailsapp/TransferBindings";
import type { TransferProgress } from "@/types/domain";

// Hit window.go for methods that may not yet be in the generated typings
// (wails dev only regenerates .d.ts on cold start).
function transferBindings() {
  const ns = (
    window as unknown as {
      go?: {
        wailsapp?: {
          TransferBindings?: {
            DownloadFolder: (remoteRoot: string, localRoot: string) => Promise<string>;
            DownloadFolderAsZip: (remoteRoot: string, localZip: string) => Promise<string>;
          };
        };
      };
    }
  ).go?.wailsapp?.TransferBindings;
  if (!ns) throw new Error("TransferBindings not available");
  return ns;
}

export const transferApi = {
  upload: (localPath: string, remotePath: string): Promise<string> =>
    UploadFile(localPath, remotePath),
  download: (remotePath: string, localPath: string): Promise<string> =>
    DownloadFile(remotePath, localPath),
  downloadFolder: (remoteRoot: string, localRoot: string): Promise<string> =>
    transferBindings().DownloadFolder(remoteRoot, localRoot),
  downloadFolderAsZip: (remoteRoot: string, localZip: string): Promise<string> =>
    transferBindings().DownloadFolderAsZip(remoteRoot, localZip),
  cancel: (id: string): Promise<void> => CancelTransfer(id),
  list: (): Promise<TransferProgress[]> =>
    ListTransfers() as Promise<TransferProgress[]>,
};
