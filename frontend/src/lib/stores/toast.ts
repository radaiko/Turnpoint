import { writable } from "svelte/store";

export interface Toast {
  id: number;
  kind: "info" | "ok" | "warn" | "danger";
  message: string;
}

export const toasts = writable<Toast[]>([]);
let nextId = 1;

export function toast(message: string, kind: Toast["kind"] = "info") {
  const id = nextId++;
  toasts.update((t) => [...t, { id, kind, message }]);
  setTimeout(() => toasts.update((t) => t.filter((x) => x.id !== id)), 4000);
}

export function dismiss(id: number) {
  toasts.update((t) => t.filter((x) => x.id !== id));
}
