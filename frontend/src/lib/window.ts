// Window controls. The app runs frameless (no OS chrome), so the custom titlebar
// provides minimise / maximise / close. Guarded so the browser dev harness (no
// Wails runtime) is a no-op.
import { WindowMinimise, WindowToggleMaximise, Quit } from "$wails/runtime/runtime";

const hasRuntime = () => typeof window !== "undefined" && !!(window as any).runtime;

export function minimise() {
  if (hasRuntime()) WindowMinimise();
}
export function toggleMaximise() {
  if (hasRuntime()) WindowToggleMaximise();
}
export function close() {
  if (hasRuntime()) Quit();
}
