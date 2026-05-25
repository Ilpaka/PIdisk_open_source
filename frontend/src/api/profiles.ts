import {
  ActiveProfile,
  ClearActiveProfile,
  CreateProfile,
  DeleteProfile,
  GetProfile,
  HasPassphrase,
  ListProfiles,
  SetActiveProfile,
  SuggestDefaults,
} from "../../wailsjs/go/wailsapp/ProfileBindings";
import type { Profile, ProfileInput } from "@/types/domain";

export interface ActiveProfileResult {
  profile: Profile;
  present: boolean;
}

export interface ProfileDefaults {
  privateKeyPath: string;
  remoteRoot: string;
  remoteTrash: string;
  localSyncDir: string;
}

export interface CreateProfileResult {
  profile: Profile;
  generatedPublicKey?: string;
  generatedKeyPath?: string;
}

export const profilesApi = {
  list: (): Promise<Profile[]> => ListProfiles() as Promise<Profile[]>,
  get: (id: string): Promise<Profile> => GetProfile(id) as Promise<Profile>,
  create: (input: ProfileInput): Promise<CreateProfileResult> =>
    CreateProfile(input as unknown as Parameters<typeof CreateProfile>[0]) as unknown as Promise<CreateProfileResult>,
  delete: (id: string): Promise<void> => DeleteProfile(id),
  setActive: (id: string): Promise<Profile> =>
    SetActiveProfile(id) as Promise<Profile>,
  clearActive: () => ClearActiveProfile(),
  active: (): Promise<ActiveProfileResult> =>
    ActiveProfile() as Promise<ActiveProfileResult>,
  hasPassphrase: (id: string): Promise<boolean> => HasPassphrase(id),
  suggestDefaults: (name: string, username: string): Promise<ProfileDefaults> =>
    SuggestDefaults(name, username) as Promise<ProfileDefaults>,
};
