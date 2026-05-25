import {
  ConfirmHostKey,
  IsConnected,
  Lock,
  NewTrashID,
  Unlock,
} from "../../wailsjs/go/wailsapp/ConnectionBindings";
import type { Profile } from "@/types/domain";

export interface UnlockResult {
  profile: Profile;
  connected: boolean;
}

export const connectionApi = {
  unlock: (profileId: string, passphrase: string): Promise<UnlockResult> =>
    Unlock(profileId, passphrase) as Promise<UnlockResult>,
  lock: (): Promise<void> => Lock(),
  isConnected: (): Promise<boolean> => IsConnected(),
  confirmHostKey: (fingerprint: string, accept: boolean): Promise<void> =>
    ConfirmHostKey(fingerprint, accept),
  newTrashId: (): Promise<string> => NewTrashID(),
};
