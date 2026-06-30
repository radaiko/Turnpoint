import { writable } from "svelte/store";
import { App, type UpdateInfo } from "$lib/api";

export interface UpdateState {
  info: UpdateInfo | null;
  checking: boolean;
  dismissed: boolean;
  installing: boolean;
}

export const update = writable<UpdateState>({
  info: null,
  checking: false,
  dismissed: false,
  installing: false,
});

// checkForUpdate asks GitHub for the latest release (best-effort; the app's only
// network call). Returns the info, or null on failure.
export async function checkForUpdate(): Promise<UpdateInfo | null> {
  update.update((s) => ({ ...s, checking: true }));
  try {
    const info = await App.CheckForUpdate();
    update.update((s) => ({ ...s, info, checking: false }));
    return info;
  } catch {
    update.update((s) => ({ ...s, checking: false }));
    return null;
  }
}

export function dismissUpdate() {
  update.update((s) => ({ ...s, dismissed: true }));
}

export function setInstalling(installing: boolean) {
  update.update((s) => ({ ...s, installing }));
}
