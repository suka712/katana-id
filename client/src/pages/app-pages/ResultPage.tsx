import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  ArrowLeftIcon,
  CheckIcon,
  XIcon,
  AlertTriangleIcon,
  DownloadIcon,
  ShieldCheckIcon,
} from "lucide-react";
import { axiosInstance } from "@/lib/axios";
import { collectFingerprint } from "@/lib/fingerprint";
import Logo from "@/components/Logo";
import { QueryInput } from "@/components/ui/QueryInput";

// ─── Platform definitions ─────────────────────────────────────────────────────

const DOMAIN_TLDS = ["com", "io", "dev", "co", "net", "org", "app", "ai", "xyz", "me"];

const CODE_PLATFORMS = [
  { key: "github", label: "GitHub" },
  { key: "gitlab", label: "GitLab" },
  { key: "npm", label: "npm" },
  { key: "pypi", label: "PyPI" },
  { key: "crates", label: "crates.io" },
  { key: "rubygems", label: "RubyGems" },
  { key: "dockerhub", label: "Docker Hub" },
  { key: "homebrew", label: "Homebrew" },
];

const COMMUNITY_PLATFORMS = [
  { key: "reddit", label: "Reddit" },
  { key: "devto", label: "dev.to" },
  { key: "keybase", label: "Keybase" },
  { key: "x", label: "X" },
];

const KNOWN_PLATFORMS = [
  ...DOMAIN_TLDS.map((t) => `domain.${t}`),
  ...CODE_PLATFORMS.map((s) => s.key),
  ...COMMUNITY_PLATFORMS.map((s) => s.key),
];

// ─── Types ────────────────────────────────────────────────────────────────────

type CellState = "pending" | "available" | "taken" | "error";

interface CellData {
  state: CellState;
  meta?: Record<string, string>;
}

// name → platform → cell
type ResultsMap = Record<string, Record<string, CellData | undefined>>;

interface SSEPayload {
  name?: string;
  platform?: string;
  available?: boolean;
  meta?: Record<string, string> | null;
  err?: string;
  done?: boolean;
}

interface Color {
  name: string;
  hex: string;
}

interface Concept {
  names: string[];
  tagline: string;
  mission: string;
  palette: Color[];
  keywords: string[];
}

interface Trust {
  value: number;
  level: string;
  reasons: string[];
}

interface GenerateResponse {
  id: string;
  concept: Concept;
  names: string[];
  total: number;
  trust: Trust;
}

type Status = "starting" | "streaming" | "done" | "error";

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function ResultPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const query = searchParams.get("q") ?? "";

  const [status, setStatus] = useState<Status>("starting");
  const [errorMsg, setErrorMsg] = useState("");
  const [kitId, setKitId] = useState("");
  const [concept, setConcept] = useState<Concept | null>(null);
  const [trust, setTrust] = useState<Trust | null>(null);
  const [names, setNames] = useState<string[]>([]);
  const [results, setResults] = useState<ResultsMap>({});
  const [total, setTotal] = useState(0);
  const [completed, setCompleted] = useState(0);
  const [elapsed, setElapsed] = useState(0);

  const startRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!query) {
      navigate("/");
      return;
    }

    setStatus("starting");
    setErrorMsg("");
    setKitId("");
    setConcept(null);
    setTrust(null);
    setNames([]);
    setResults({});
    setTotal(0);
    setCompleted(0);
    setElapsed(0);

    const run = async () => {
      try {
        const fp = await collectFingerprint();

        const res = await axiosInstance.post<GenerateResponse>("/generate", {
          prompt: query,
          fingerprint: fp.fingerprint,
          components: fp.components,
        });

        const { id, concept, names, total, trust } = res.data;

        setKitId(id);
        setConcept(concept);
        setTrust(trust);
        setNames(names);
        setTotal(total);
        setStatus("streaming");

        // Pre-seed platforms as pending so pills appear immediately.
        const skeleton: ResultsMap = {};
        for (const n of names) {
          skeleton[n] = Object.fromEntries(KNOWN_PLATFORMS.map((p) => [p, undefined]));
        }
        setResults(skeleton);

        startRef.current = Date.now();
        timerRef.current = setInterval(
          () => setElapsed(Date.now() - startRef.current),
          100
        );

        const es = new EventSource(
          `${import.meta.env.VITE_API_URL}/generate/${id}/stream`,
          { withCredentials: true }
        );
        esRef.current = es;

        es.onmessage = (e: MessageEvent) => {
          const payload: SSEPayload = JSON.parse(e.data);

          if (payload.done) {
            clearInterval(timerRef.current!);
            setElapsed(Date.now() - startRef.current);
            setStatus("done");
            es.close();
            return;
          }

          const { name, platform, available, meta, err } = payload;
          if (!name || !platform) return;

          const state: CellState = err ? "error" : available ? "available" : "taken";

          setResults((prev) => ({
            ...prev,
            [name]: {
              ...(prev[name] ?? {}),
              [platform]: { state, meta: meta ?? undefined },
            },
          }));
          setCompleted((n) => n + 1);
        };

        es.onerror = () => {
          clearInterval(timerRef.current!);
          es.close();
          setStatus((s) => (s === "streaming" ? "done" : s));
        };
      } catch (err: any) {
        if (err?.response?.status === 401) {
          navigate(
            `/signin?redirect=${encodeURIComponent(
              `/result?q=${encodeURIComponent(query)}`
            )}`
          );
          return;
        }
        setErrorMsg(err?.response?.data?.error ?? "Something went wrong");
        setStatus("error");
      }
    };

    run();

    return () => {
      esRef.current?.close();
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [query]);

  return (
    <div className="min-h-screen flex flex-col bg-background">
      {/* ── Header ── */}
      <header className="sticky top-0 z-40 border-b border-border/30 bg-background/80 backdrop-blur-md">
        <div className="max-w-5xl mx-auto px-6 py-3 flex items-center justify-between gap-6">
          <div className="flex items-center gap-4">
            <Link
              to="/"
              className="p-1.5 -ml-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-white/5 transition-colors"
            >
              <ArrowLeftIcon className="w-4 h-4" />
            </Link>
            <div className="flex items-center gap-2 opacity-60 hover:opacity-100 transition-opacity">
              <Logo />
              <span className="text-sm font-medium">KatanaID</span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {trust && <TrustBadge trust={trust} />}
            <StatusDisplay
              status={status}
              completed={completed}
              total={total}
              elapsed={elapsed}
            />
          </div>
        </div>
      </header>

      {/* ── Body ── */}
      <main className="flex-1 flex flex-col items-center px-6 py-12 max-w-5xl mx-auto w-full">
        <motion.div
          className="w-full mb-10"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ type: "spring", stiffness: 380, damping: 32 }}
        >
          <QueryInput
            key={query}
            defaultValue={query}
            onSearch={(q) => navigate(`/result?q=${encodeURIComponent(q)}`)}
            loading={status === "starting" || status === "streaming"}
            scanning={status === "streaming"}
          />
        </motion.div>

        {/* Starting spinner */}
        <AnimatePresence>
          {status === "starting" && (
            <motion.div
              key="spinner"
              className="flex flex-col items-center gap-4 py-24"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
            >
              <ScanningRing />
              <p className="text-sm text-muted-foreground/60 tracking-widest uppercase">
                Generating brand…
              </p>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Error */}
        {status === "error" && (
          <motion.div
            className="flex flex-col items-center gap-4 py-24"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
            <p className="text-muted-foreground">{errorMsg}</p>
            <Link to="/" className="text-sm text-primary hover:underline">
              ← Try again
            </Link>
          </motion.div>
        )}

        {/* Concept */}
        {concept && (
          <ConceptHeader
            concept={concept}
            kitId={kitId}
            canDownload={status === "done"}
          />
        )}

        {/* Name cards */}
        <div className="w-full flex flex-col gap-6">
          {names.map((name, i) => (
            <NameCard key={name} name={name} index={i} cells={results[name] ?? {}} />
          ))}
        </div>
      </main>
    </div>
  );
}

// ─── Concept header ───────────────────────────────────────────────────────────

function ConceptHeader({
  concept,
  kitId,
  canDownload,
}: {
  concept: Concept;
  kitId: string;
  canDownload: boolean;
}) {
  return (
    <motion.div
      className="w-full mb-10 rounded-2xl border border-border/40 bg-card/60 backdrop-blur-sm p-8 md:p-10"
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
    >
      <div className="flex items-start justify-between gap-6 flex-wrap">
        <div className="min-w-0">
          <p className="text-[10px] uppercase tracking-[0.25em] text-muted-foreground/40 mb-2">
            Brand Concept
          </p>
          {concept.tagline && (
            <h1 className="font-heading text-3xl md:text-4xl leading-tight bg-linear-to-r from-violet-400 via-indigo-400 to-cyan-400 bg-clip-text text-transparent">
              {concept.tagline}
            </h1>
          )}
          {concept.mission && (
            <p className="mt-3 text-sm text-muted-foreground/70 max-w-xl">
              {concept.mission}
            </p>
          )}
        </div>

        <a
          href={
            canDownload
              ? `${import.meta.env.VITE_API_URL}/kits/${kitId}/pdf`
              : undefined
          }
          aria-disabled={!canDownload}
          className={`shrink-0 inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium ring-1 transition-colors ${
            canDownload
              ? "bg-primary/10 text-primary ring-primary/30 hover:bg-primary/20"
              : "bg-muted/20 text-muted-foreground/40 ring-white/5 pointer-events-none"
          }`}
        >
          <DownloadIcon className="w-4 h-4" />
          {canDownload ? "Download report" : "Preparing report…"}
        </a>
      </div>

      {/* Palette */}
      {concept.palette?.length > 0 && (
        <div className="mt-8 flex flex-wrap gap-3">
          {concept.palette.map((c) => (
            <div key={c.hex} className="flex flex-col gap-1.5">
              <div
                className="w-16 h-16 rounded-xl ring-1 ring-white/10"
                style={{ backgroundColor: c.hex }}
              />
              <span className="text-[10px] text-muted-foreground/60">{c.name}</span>
              <span className="text-[10px] font-mono text-muted-foreground/40 uppercase">
                {c.hex}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Keywords */}
      {concept.keywords?.length > 0 && (
        <div className="mt-6 flex flex-wrap gap-2">
          {concept.keywords.map((k) => (
            <span
              key={k}
              className="rounded-full bg-white/5 px-3 py-1 text-xs text-muted-foreground/70 ring-1 ring-white/5"
            >
              {k}
            </span>
          ))}
        </div>
      )}
    </motion.div>
  );
}

// ─── Trust badge ──────────────────────────────────────────────────────────────

function TrustBadge({ trust }: { trust: Trust }) {
  const color =
    trust.level === "high"
      ? "text-cyan-400 ring-cyan-400/30 bg-cyan-500/10"
      : trust.level === "medium"
      ? "text-amber-400 ring-amber-400/30 bg-amber-500/10"
      : "text-red-400 ring-red-400/30 bg-red-500/10";

  return (
    <span
      title={trust.reasons.join(" · ")}
      className={`hidden sm:inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-mono ring-1 ${color}`}
    >
      <ShieldCheckIcon className="w-3.5 h-3.5" />
      trust {trust.value}
    </span>
  );
}

// ─── NameCard ─────────────────────────────────────────────────────────────────

function NameCard({
  name,
  index,
  cells,
}: {
  name: string;
  index: number;
  cells: Record<string, CellData | undefined>;
}) {
  const availableCount = Object.values(cells).filter(
    (c) => c?.state === "available"
  ).length;
  const resolvedCount = Object.values(cells).filter(
    (c) => c?.state !== undefined
  ).length;

  const scoreRatio = resolvedCount > 0 ? availableCount / resolvedCount : 0;
  const scoreColor =
    scoreRatio > 0.66
      ? "text-cyan-400"
      : scoreRatio > 0.33
      ? "text-amber-400"
      : "text-red-400";

  return (
    <motion.div
      initial={{ opacity: 0, y: 24 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: index * 0.08, ease: [0.16, 1, 0.3, 1] }}
    >
      <div className="relative rounded-2xl border border-border/40 bg-card/60 backdrop-blur-sm overflow-hidden">
        <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r from-transparent via-primary/40 to-transparent" />

        <div className="relative p-8 md:p-10">
          <div className="flex items-start justify-between gap-4 mb-8">
            <div>
              <p className="text-[10px] uppercase tracking-[0.25em] text-muted-foreground/40 mb-1">
                Identity
              </p>
              <h2 className="font-heading text-5xl md:text-6xl italic leading-none bg-linear-to-r from-violet-400 via-indigo-400 to-cyan-400 bg-clip-text text-transparent">
                {name}
              </h2>
            </div>

            <div className="flex flex-col items-end gap-1 pt-1 shrink-0">
              <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/40">
                Available
              </p>
              <p className={`font-mono text-2xl font-bold tabular-nums ${scoreColor}`}>
                {availableCount}
                <span className="text-muted-foreground/30 text-base font-normal">
                  /{resolvedCount}
                </span>
              </p>
            </div>
          </div>

          <div className="w-full h-px bg-border/30 mb-8" />

          {/* Domains */}
          <section className="mb-8">
            <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/40 mb-4">
              Domains
            </p>
            <div className="flex flex-wrap gap-2">
              {DOMAIN_TLDS.map((tld) => (
                <LED key={tld} label={`.${tld}`} data={cells[`domain.${tld}`]} />
              ))}
            </div>
          </section>

          {/* Code & packages */}
          <section className="mb-8">
            <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/40 mb-4">
              Code & Packages
            </p>
            <div className="flex flex-wrap gap-3">
              {CODE_PLATFORMS.map((p) => (
                <LED key={p.key} label={p.label} data={cells[p.key]} large />
              ))}
            </div>
          </section>

          {/* Community */}
          <section>
            <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/40 mb-4">
              Community
            </p>
            <div className="flex flex-wrap gap-3">
              {COMMUNITY_PLATFORMS.map((p) => (
                <LED key={p.key} label={p.label} data={cells[p.key]} large />
              ))}
            </div>
          </section>
        </div>
      </div>
    </motion.div>
  );
}

// ─── LED pill ─────────────────────────────────────────────────────────────────

function LED({
  label,
  data,
  large = false,
}: {
  label: string;
  data: CellData | undefined;
  large?: boolean;
}) {
  const state = data?.state;

  const styles: Record<string, string> = {
    pending: "bg-muted/20 text-muted-foreground/25 ring-1 ring-white/5",
    available: "bg-cyan-500/10 text-cyan-300 ring-1 ring-cyan-400/30",
    taken: "bg-red-950/30 text-red-400/50 ring-1 ring-red-900/40",
    error: "bg-amber-950/20 text-amber-500/50 ring-1 ring-amber-800/20",
  };

  const dotStyles: Record<string, string> = {
    pending: "bg-white/10 animate-pulse",
    available: "bg-cyan-400 shadow-[0_0_6px_1px_oklch(75%_0.18_195/0.8)]",
    taken: "bg-red-500/50",
    error: "bg-amber-500/50",
  };

  const resolvedState = state ?? "pending";

  return (
    <motion.div
      layout
      initial={false}
      animate={
        state && state !== "pending"
          ? { scale: [1, 1.08, 1], transition: { duration: 0.3, ease: "easeOut" } }
          : {}
      }
      className={`
        flex items-center gap-2 font-mono
        rounded-full transition-colors duration-300
        ${large ? "px-4 py-2 text-sm" : "px-3 py-1.5 text-xs"}
        ${styles[resolvedState]}
      `}
    >
      <span
        className={`shrink-0 rounded-full transition-all duration-300 ${
          large ? "w-2 h-2" : "w-1.5 h-1.5"
        } ${dotStyles[resolvedState]}`}
      />
      {label}
      {state === "available" && (
        <CheckIcon className={`shrink-0 text-cyan-400/70 ${large ? "w-3.5 h-3.5" : "w-3 h-3"}`} />
      )}
      {state === "taken" && (
        <XIcon className={`shrink-0 text-red-400/40 ${large ? "w-3.5 h-3.5" : "w-3 h-3"}`} />
      )}
      {state === "error" && (
        <AlertTriangleIcon
          className={`shrink-0 text-amber-400/50 ${large ? "w-3.5 h-3.5" : "w-3 h-3"}`}
        />
      )}
    </motion.div>
  );
}

// ─── Status display ───────────────────────────────────────────────────────────

function StatusDisplay({
  status,
  completed,
  total,
  elapsed,
}: {
  status: Status;
  completed: number;
  total: number;
  elapsed: number;
}) {
  if (status === "starting") {
    return (
      <span className="flex items-center gap-2 text-xs text-muted-foreground/50 font-mono uppercase tracking-widest">
        <span className="inline-block w-1.5 h-1.5 rounded-full bg-primary/60 animate-pulse" />
        Generating
      </span>
    );
  }
  if (status === "streaming") {
    return (
      <span className="flex items-center gap-3 text-xs font-mono text-muted-foreground/60">
        <span className="flex items-center gap-1.5">
          <span className="inline-block w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse shadow-[0_0_6px_oklch(75%_0.18_195)]" />
          <span className="text-foreground/80">{completed}</span>
          <span className="text-muted-foreground/30">/</span>
          <span>{total}</span>
        </span>
        <span className="text-muted-foreground/20">·</span>
        <span className="tabular-nums">{(elapsed / 1000).toFixed(1)}s</span>
      </span>
    );
  }
  if (status === "done") {
    return (
      <span className="flex items-center gap-2 text-xs font-mono">
        <span className="inline-flex items-center justify-center w-4 h-4 rounded-full bg-cyan-500/15 ring-1 ring-cyan-400/30">
          <CheckIcon className="w-2.5 h-2.5 text-cyan-400" />
        </span>
        <span className="text-muted-foreground/60">
          <span className="text-foreground/80 tabular-nums">{completed}</span> checks
        </span>
        <span className="text-muted-foreground/20">·</span>
        <span className="text-foreground/60 tabular-nums">{(elapsed / 1000).toFixed(1)}s</span>
      </span>
    );
  }
  return null;
}

// ─── Scanning ring (starting state) ──────────────────────────────────────────

function ScanningRing() {
  return (
    <div className="relative w-14 h-14">
      <div className="absolute inset-0 rounded-full border border-border/30" />
      <motion.div
        className="absolute inset-0 rounded-full border-2 border-transparent"
        style={{
          borderTopColor: "oklch(65% 0.22 268)",
          borderRightColor: "oklch(65% 0.22 268 / 0.3)",
        }}
        animate={{ rotate: 360 }}
        transition={{ repeat: Infinity, duration: 1.2, ease: "linear" }}
      />
      <div className="absolute inset-0 flex items-center justify-center">
        <div className="w-1.5 h-1.5 rounded-full bg-primary/60 animate-pulse" />
      </div>
    </div>
  );
}
