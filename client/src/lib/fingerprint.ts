// Browser fingerprinting for the trust-score engine. Collects stable, passive
// device/browser signals and hashes them into a single fingerprint the backend
// scores. Nothing here is personally identifying — it is entropy about the
// environment, used to tell real browsers from automation.

export interface Fingerprint {
  fingerprint: string;
  components: Record<string, string>;
}

async function sha256Hex(input: string): Promise<string> {
  const bytes = new TextEncoder().encode(input);
  const digest = await crypto.subtle.digest("SHA-256", bytes);
  return [...new Uint8Array(digest)]
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

// canvasSignature renders text + shapes and hashes the pixels. GPU/driver/font
// differences make this stable per-device but varied across devices, and it is
// typically absent/uniform in headless environments.
function canvasSignature(): string {
  try {
    const canvas = document.createElement("canvas");
    canvas.width = 240;
    canvas.height = 60;
    const ctx = canvas.getContext("2d");
    if (!ctx) return "no-canvas";
    ctx.textBaseline = "top";
    ctx.font = "16px 'Arial'";
    ctx.fillStyle = "#f60";
    ctx.fillRect(10, 10, 80, 30);
    ctx.fillStyle = "#069";
    ctx.fillText("KatanaID ✨ trust", 12, 14);
    ctx.strokeStyle = "rgba(0,120,200,0.6)";
    ctx.arc(60, 30, 20, 0, Math.PI * 2);
    ctx.stroke();
    const data = canvas.toDataURL();
    // Compress the long data URL into a short djb2 hash.
    let h = 5381;
    for (let i = 0; i < data.length; i++) h = (h * 33) ^ data.charCodeAt(i);
    return (h >>> 0).toString(16);
  } catch {
    return "canvas-error";
  }
}

export async function collectFingerprint(): Promise<Fingerprint> {
  const nav = navigator as Navigator & {
    deviceMemory?: number;
    hardwareConcurrency?: number;
  };

  const components: Record<string, string> = {
    lang: nav.language ?? "",
    languages: (nav.languages ?? []).join(","),
    platform: nav.platform ?? "",
    tz: Intl.DateTimeFormat().resolvedOptions().timeZone ?? "",
    tzOffset: String(new Date().getTimezoneOffset()),
    screen: `${screen.width}x${screen.height}x${screen.colorDepth}`,
    viewport: `${window.innerWidth}x${window.innerHeight}`,
    dpr: String(window.devicePixelRatio ?? 1),
    cores: String(nav.hardwareConcurrency ?? ""),
    memory: String(nav.deviceMemory ?? ""),
    touch: String(nav.maxTouchPoints ?? 0),
    canvas: canvasSignature(),
  };

  const fingerprint = await sha256Hex(Object.values(components).join("|"));
  return { fingerprint, components };
}
