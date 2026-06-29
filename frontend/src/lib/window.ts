// Window controls. The app runs frameless (no OS chrome), so the custom titlebar
// provides minimise / maximise / close. Guarded so the browser dev harness (no
// Wails runtime) is a no-op.
import { WindowMinimise, WindowToggleMaximise, Quit, Environment } from "$wails/runtime/runtime";

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

export type Platform = "darwin" | "windows" | "linux";

// getPlatform resolves the host OS so the titlebar can place window controls on
// the platform-appropriate side (macOS left, Windows/Linux right). Falls back to
// the user agent in the browser dev harness.
export async function getPlatform(): Promise<Platform> {
  if (hasRuntime()) {
    try {
      const env = await Environment();
      if (env.platform === "darwin" || env.platform === "windows" || env.platform === "linux") {
        return env.platform;
      }
    } catch {
      /* fall through */
    }
  }
  const ua = typeof navigator !== "undefined" ? navigator.userAgent : "";
  if (/Mac/i.test(ua)) return "darwin";
  if (/Win/i.test(ua)) return "windows";
  return "linux";
}
