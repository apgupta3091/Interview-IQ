import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { RefreshCw } from "lucide-react";
import { api } from "@/lib/api";
import type { Problem, CategoryStats } from "@/types/api";

const PANEL_LIMIT = 8;

function difficultyColor(d: string | undefined) {
  if (d === "easy") return "text-green-500";
  if (d === "hard") return "text-red-500";
  return "text-amber-500";
}

function scoreColor(s: number) {
  if (s >= 70)
    return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
  if (s >= 40)
    return "bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400";
  return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
}

type RankedProblem = {
  problem: Problem;
  priority: number;
};

export default function RetryPanel() {
  const [items, setItems] = useState<Problem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const [problemsResp, statsResp] = await Promise.all([
          api.problems.listFiltered({ limit: 200 }),
          api.categories.stats(),
        ]);
        if (cancelled) return;

        // Deduplicate by problem name, keeping the latest submission (results are
        // already sorted by created_at DESC, so the first occurrence is the latest).
        const seen = new Set<string>();
        const problems = (problemsResp.problems ?? []).filter((p) => {
          if (seen.has(p.name)) return false;
          seen.add(p.name);
          return true;
        });
        const stats: CategoryStats[] = statsResp;

        // Build category -> strength lookup
        const strengthMap = new Map<string, number>();
        stats.forEach((s) => {
          if (s.category && s.strength !== undefined) {
            strengthMap.set(s.category, s.strength);
          }
        });

        // Only show problems the user struggled with (score < 80).
        // Priority = (100 - score) * (100 - weakest_category_strength) / 100
        // Higher priority = worse score in a weaker category → most urgent to retry.
        const ranked: RankedProblem[] = problems
          .filter((p) => (p.score ?? 100) < 80)
          .map((p) => {
            const score = p.score ?? 100;
            const cats = p.categories ?? [];
            const minStrength =
              cats.length > 0
                ? Math.min(...cats.map((c) => strengthMap.get(c) ?? 100))
                : 100;
            const priority = ((100 - score) * (100 - minStrength)) / 100;
            return { problem: p, priority };
          })
          .sort((a, b) => b.priority - a.priority)
          .slice(0, PANEL_LIMIT);

        setItems(ranked.map((r) => r.problem));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <aside className="fixed right-0 top-0 h-screen w-56 border-l bg-sidebar flex flex-col z-30">
      {/* Header */}
      <div className="px-4 py-[1.0625rem] border-b border-sidebar-border flex items-center gap-2 shrink-0">
        <RefreshCw className="w-4 h-4 text-primary shrink-0" />
        <span className="font-semibold text-sm tracking-tight">Retry List</span>
      </div>

      {/* Problem list */}
      <div className="flex-1 overflow-y-auto">
        {loading && (
          <p className="text-xs text-muted-foreground px-4 py-4">Loading…</p>
        )}

        {!loading && items.length === 0 && (
          <p className="text-xs text-muted-foreground px-4 py-4 leading-relaxed">
            No problems to retry yet — keep solving!
          </p>
        )}

        {!loading &&
          items.map((p) => (
            <Link
              key={p.id}
              to={`/problems/${p.id}`}
              className="block px-4 py-3 border-b border-sidebar-border/50 hover:bg-sidebar-accent group transition-colors"
            >
              {/* Problem name */}
              <p className="text-xs font-medium text-sidebar-foreground group-hover:text-primary truncate leading-snug">
                {p.name}
              </p>

              {/* Difficulty + score badge */}
              <div className="flex items-center gap-1.5 mt-1.5">
                <span
                  className={`text-[10px] font-medium capitalize leading-none ${difficultyColor(p.difficulty)}`}
                >
                  {p.difficulty}
                </span>
                <span
                  className={`text-[10px] font-medium px-1.5 py-0.5 rounded-full leading-none ${scoreColor(p.score ?? 0)}`}
                >
                  {p.score}
                </span>
              </div>

              {/* Categories */}
              {p.categories && p.categories.length > 0 && (
                <p className="text-[10px] text-muted-foreground mt-1 truncate">
                  {p.categories.slice(0, 2).join(", ")}
                  {p.categories.length > 2 ? "…" : ""}
                </p>
              )}
            </Link>
          ))}
      </div>

      {/* Footer hint */}
      {!loading && items.length > 0 && (
        <div className="px-4 py-3 border-t border-sidebar-border shrink-0">
          <p className="text-[10px] text-muted-foreground leading-relaxed">
            Sorted by score × category weakness
          </p>
        </div>
      )}
    </aside>
  );
}
