import {
  CancelTransfer,
  DownloadFile,
  ListTransfers,
  UploadFile,
} from "../../wailsjs/go/wailsapp/TransferBindings";
import type { TransferProgress } from "@/types/domain";

export const transferApi = {
  upload: (localPath: string, remotePath: string): Promise<string> =>
    UploadFile(localPath, remotePath),
  download: (remotePath: string, localPath: string): Promise<string> =>
    DownloadFile(remotePath, localPath),
  cancel: (id: string): Promise<void> => CancelTransfer(id),
  list: (): Promise<TransferProgress[]> =>
    ListTransfers() as Promise<TransferProgress[]>,
};
