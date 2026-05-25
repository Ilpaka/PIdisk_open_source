import { create } from "zustand";
import type { Profile } from "@/types/domain";
import { profilesApi, type CreateProfileResult } from "@/api/profiles";

export type ProfileStatus = "loading" | "none" | "selected" | "active";

interface ProfileState {
  status: ProfileStatus;
  profiles: Profile[];
  active: Profile | null;
  error: string | null;
  refresh: () => Promise<void>;
  create: (
    input: Parameters<typeof profilesApi.create>[0],
  ) => Promise<CreateProfileResult>;
  remove: (id: string) => Promise<void>;
  setActive: (id: string) => Promise<Profile>;
  markActive: (profile: Profile) => void;
  clearActive: () => Promise<void>;
}

export const useProfileStore = create<ProfileState>((set, get) => ({
  status: "loading",
  profiles: [],
  active: null,
  error: null,

  async refresh() {
    set({ status: "loading", error: null });
    try {
      const list = await profilesApi.list();
      const status: ProfileStatus = list.length === 0 ? "none" : get().active ? "active" : "selected";
      set({ profiles: list, status });
    } catch (err) {
      set({ error: errMessage(err), status: "none", profiles: [] });
    }
  },

  async create(input) {
    const result = await profilesApi.create(input);
    await get().refresh();
    return result;
  },

  async remove(id) {
    await profilesApi.delete(id);
    if (get().active?.id === id) {
      set({ active: null });
    }
    await get().refresh();
  },

  async setActive(id) {
    const p = await profilesApi.setActive(id);
    set({ active: p, status: "active" });
    return p;
  },

  // markActive is the local-only counterpart to setActive. Use it after the
  // connection layer has already acknowledged the profile (it sets the active
  // profile backend-side as part of Connect), to avoid a redundant round trip.
  markActive(profile) {
    set({ active: profile, status: "active" });
  },

  async clearActive() {
    await profilesApi.clearActive();
    set({ active: null, status: get().profiles.length === 0 ? "none" : "selected" });
  },
}));

function errMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  return String(err);
}
