import { useState } from "react";
import { Loader2Icon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface QueryInputProps {
  defaultValue?: string;
  onSearch: (query: string) => void;
  /** Disables input + shows spinner on button while a check is running */
  loading?: boolean;
  /** Pulses the glow to match the scanning cards */
  scanning?: boolean;
  className?: string;
}

export function QueryInput({
  defaultValue = "",
  onSearch,
  loading = false,
  scanning = false,
  className,
}: QueryInputProps) {
  const [value, setValue] = useState(defaultValue);

  const submit = () => {
    const q = value.trim();
    if (!q || loading) return;
    onSearch(q);
  };

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <Input
        placeholder='I am building Tinder but for Dog lovers called "Ruffle" . . .'
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && submit()}
        disabled={loading}
        className={cn(
          "flex-1 rounded-full border-white/10 bg-white/5 backdrop-blur-sm",
          "focus-visible:border-primary/40",
          "transition-[box-shadow] duration-700",
          scanning
            ? "shadow-[0_0_32px_-4px_oklch(65%_0.22_268/0.55)]"
            : "shadow-[0_0_24px_-6px_oklch(65%_0.22_268/0.25)]",
        )}
      />
      <Button
        onClick={submit}
        disabled={loading}
        className={cn(
          "shrink-0 rounded-full",
          "shadow-[0_0_24px_-4px_oklch(65%_0.22_268/0.5)]",
          "hover:shadow-[0_0_36px_-4px_oklch(65%_0.22_268/0.75)]",
          "transition-shadow duration-300",
        )}
      >
        {loading
          ? <Loader2Icon className="w-4 h-4 animate-spin" />
          : "Check"
        }
      </Button>
    </div>
  );
}
