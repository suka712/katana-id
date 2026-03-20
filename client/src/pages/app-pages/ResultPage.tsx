import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import { ArrowLeftIcon, CheckIcon, XIcon, AlertTriangleIcon } from "lucide-react";
import { axiosInstance } from "@/lib/axios";
import Logo from "@/components/Logo";
import { QueryInput } from "@/components/ui/QueryInput";

// ─── Platform definitions ─────────────────────────────────────────────────────

const DOMAIN_TLDS = ["com", "io", "dev", "co", "net", "org", "app", "ai", "xyz", "me"];

const SOCIAL_PLATFORMS = [
  { key: "github",  label: "GitHub" },
  { key: "npm",     label: "npm"    },
  { key: "reddit",  label: "Reddit" },
];

const KNOWN_PLATFORMS = [
  ...DOMAIN_TLDS.map((t) => `domain.${t}`),
  ...SOCIAL_PLATFORMS.map((s) => s.key),
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
  Name?: string;
  Platform?: string;
  Available?: boolean;
  Meta?: Record<string, string> | null;
  Err?: string;
  done?: boolean;
}

type Status = "starting" | "streaming" | "done" | "error";

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function ResultPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const query = searchParams.get("q") ?? "";

  const [status, setStatus]       = useState<Status>("starting");
  const [errorMsg, setErrorMsg]   = useState("");
  const [names, setNames]         = useState<string[]>([]);
  const [results, setResults]     = useState<ResultsMap>({});
  const [total, setTotal]         = useState(0);
  const [completed, setCompleted] = useState(0);
  const [elapsed, setElapsed]     = useState(0);
  const [hasSearch, setHasSearch] = useState(false);

  const startRef  = useRef(0);
  const timerRef  = useRef<ReturnType<typeof setInterval> | null>(null);
  const esRef     = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!query) { navigate("/"); return; }

    // Reset all state for each new query
    setStatus("starting");
    setErrorMsg("");
    setNames([]);
    setResults({});
    setTotal(0);
    setCompleted(0);
    setElapsed(0);
    setHasSearch(false);

    const run = async () => {
      try {
        const res = await axiosInstance.post<{
          id: string; names: string[]; total: number;
        }>("/check", { query });

        const { id, names: checkNames, total: checkTotal } = res.data;

        setNames(checkNames);
        setTotal(checkTotal);
        setStatus("streaming");

        // Pre-seed all known platforms as pending so pills appear immediately
        const skeleton: ResultsMap = {};
        for (const n of checkNames) {
          skeleton[n] = Object.fromEntries(KNOWN_PLATFORMS.map((p) => [p, undefined]));
        }
        setResults(skeleton);

        startRef.current = Date.now();
        timerRef.current = setInterval(
          () => setElapsed(Date.now() - startRef.current),
          100
        );

        const es = new EventSource(
          `${import.meta.env.VITE_API_URL}/check/${id}`,
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

          const { Name, Platform, Available, Meta, Err } = payload;
          if (!Name || !Platform) return;

          if (Platform === "search") setHasSearch(true);

          const state: CellState = Err ? "error" : Available ? "available" : "taken";

          setResults((prev) => ({
            ...prev,
            [Name]: {
              ...(prev[Name] ?? {}),
              [Platform]: { state, meta: Meta ?? undefined },
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
              window.location.pathname + window.location.search
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
          {/* Left: back + logo */}
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

          {/* Right: status */}
          <StatusDisplay
            status={status}
            completed={completed}
            total={total}
            elapsed={elapsed}
          />
        </div>
      </header>

      {/* ── Body ── */}
      <main className="flex-1 flex flex-col items-center px-6 py-12 max-w-5xl mx-auto w-full">

        {/* Search bar — springs in from above, mirroring where it just was */}
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
                Initialising checks…
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

        {/* Name cards */}
        <div className="w-full flex flex-col gap-6">
          {names.map((name, i) => (
            <NameCard
              key={name}
              name={name}
              index={i}
              cells={results[name] ?? {}}
              status={status}
              completed={completed}
              total={total}
              hasSearch={hasSearch}
            />
          ))}
        </div>
      </main>
    </div>
  );
}

// ─── NameCard ─────────────────────────────────────────────────────────────────

function NameCard({
  name,
  index,
  cells,
  // status,
  // completed,
  total,
  hasSearch,
}: {
  name: string;
  index: number;
  cells: Record<string, CellData | undefined>;
  status: Status;
  completed: number;
  total: number;
  hasSearch: boolean;
}) {
  const availableCount = Object.values(cells).filter(
    (c) => c?.state === "available"
  ).length;
  const resolvedCount = Object.values(cells).filter(
    (c) => c?.state !== undefined
  ).length;

  const scoreRatio = total > 0 ? availableCount / total : 0;
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
        {/* Top inset gradient line */}
        <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r from-transparent via-primary/40 to-transparent" />

        <div className="relative p-8 md:p-10">
          {/* Name + score */}
          <div className="flex items-start justify-between gap-4 mb-8">
            <div>
              <p className="text-[10px] uppercase tracking-[0.25em] text-muted-foreground/40 mb-1">
                Identity
              </p>
              <h2
                className="font-heading text-5xl md:text-6xl italic leading-none bg-linear-to-r from-violet-400 via-indigo-400 to-cyan-400 bg-clip-text text-transparent"
              >
                {name}
              </h2>
            </div>

            {/* Score badge */}
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

          {/* Divider */}
          <div className="w-full h-px bg-border/30 mb-8" />

          {/* Domains */}
          <section className="mb-8">
            <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/40 mb-4">
              Domains
            </p>
            <div className="flex flex-wrap gap-2">
              {DOMAIN_TLDS.map((tld) => (
                <LED
                  key={tld}
                  label={`.${tld}`}
                  data={cells[`domain.${tld}`]}
                />
              ))}
            </div>
          </section>

          {/* Social & Code */}
          <section className={hasSearch ? "mb-8" : ""}>
            <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/40 mb-4">
              Social & Code
            </p>
            <div className="flex flex-wrap gap-3">
              {SOCIAL_PLATFORMS.map((p) => (
                <LED key={p.key} label={p.label} data={cells[p.key]} large />
              ))}
            </div>
          </section>

          {/* Search presence */}
          <AnimatePresence>
            {hasSearch && cells["search"] && (
              <motion.section
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: "auto" }}
                transition={{ duration: 0.4 }}
              >
                <div className="w-full h-px bg-border/30 mb-8" />
                <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/40 mb-4">
                  Search Presence
                </p>
                <SearchBar data={cells["search"]} />
              </motion.section>
            )}
          </AnimatePresence>
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
    pending:
      "bg-muted/20 text-muted-foreground/25 ring-1 ring-white/5",
    available:
      "bg-cyan-500/10 text-cyan-300 ring-1 ring-cyan-400/30",
    taken:
      "bg-red-950/30 text-red-400/50 ring-1 ring-red-900/40",
    error:
      "bg-amber-950/20 text-amber-500/50 ring-1 ring-amber-800/20",
  };

  const dotStyles: Record<string, string> = {
    pending:   "bg-white/10 animate-pulse",
    available: "bg-cyan-400 shadow-[0_0_6px_1px_oklch(75%_0.18_195/0.8)]",
    taken:     "bg-red-500/50",
    error:     "bg-amber-500/50",
  };

  const resolvedState = state ?? "pending";

  return (
    <motion.div
      layout
      initial={false}
      animate={
        state && state !== "pending"
          ? {
              scale: [1, 1.08, 1],
              transition: { duration: 0.3, ease: "easeOut" },
            }
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
        <AlertTriangleIcon className={`shrink-0 text-amber-400/50 ${large ? "w-3.5 h-3.5" : "w-3 h-3"}`} />
      )}
    </motion.div>
  );
}

// ─── Search bar ───────────────────────────────────────────────────────────────

function SearchBar({ data }: { data: CellData | undefined }) {
  const score = data?.meta?.competitiveness ?? "unknown";
  const totalResults = data?.meta?.total_results;

  const fillMap: Record<string, number>   = { low: 0.25, medium: 0.6, high: 0.92 };
  const colorMap: Record<string, string>  = {
    low:    "from-cyan-500 to-cyan-400",
    medium: "from-amber-500 to-amber-400",
    high:   "from-red-600 to-red-500",
  };
  const labelMap: Record<string, string> = {
    low:    "Low competition",
    medium: "Moderate competition",
    high:   "High competition",
  };

  const fill  = fillMap[score]  ?? 0;
  const color = colorMap[score] ?? "from-muted to-muted";
  const label = labelMap[score] ?? "Unknown";

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between text-xs text-muted-foreground/50 font-mono">
        <span>{label}</span>
        {totalResults && <span>{Number(totalResults).toLocaleString()} results</span>}
      </div>
      <div className="h-1.5 w-full rounded-full bg-muted/20 overflow-hidden">
        <motion.div
          className={`h-full rounded-full bg-gradient-to-r ${color}`}
          initial={{ width: "0%" }}
          animate={{ width: `${fill * 100}%` }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
        />
      </div>
    </div>
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
        Initialising
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
          <span className="text-foreground/80 tabular-nums">{completed}</span>
          {" "}checks
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
      {/* Static ring */}
      <div className="absolute inset-0 rounded-full border border-border/30" />
      {/* Spinning arc */}
      <motion.div
        className="absolute inset-0 rounded-full border-2 border-transparent"
        style={{
          borderTopColor: "oklch(65% 0.22 268)",
          borderRightColor: "oklch(65% 0.22 268 / 0.3)",
        }}
        animate={{ rotate: 360 }}
        transition={{ repeat: Infinity, duration: 1.2, ease: "linear" }}
      />
      {/* Center dot */}
      <div className="absolute inset-0 flex items-center justify-center">
        <div className="w-1.5 h-1.5 rounded-full bg-primary/60 animate-pulse" />
      </div>
    </div>
  );
}
