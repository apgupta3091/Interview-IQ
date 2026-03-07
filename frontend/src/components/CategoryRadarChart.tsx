import {
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  Radar,
  ResponsiveContainer,
  Tooltip,
} from 'recharts'
import type { CategoryStats } from '@/types/api'

type Props = { stats: CategoryStats[] }

// Short label so the axis text fits around the chart
function shortLabel(cat: string) {
  const map: Record<string, string> = {
    'array': 'Array',
    'string': 'String',
    'hash-map': 'Hash Map',
    'two-pointers': '2 Pointers',
    'sliding-window': 'Sliding Win',
    'binary-search': 'Bin Search',
    'stack': 'Stack',
    'queue': 'Queue',
    'linked-list': 'Linked List',
    'tree': 'Tree',
    'trie': 'Trie',
    'heap': 'Heap',
    'graph': 'Graph',
    'advanced-graphs': 'Adv Graphs',
    'dp': 'DP',
    'dp-2d': 'DP 2D',
    'backtracking': 'Backtrack',
    'greedy': 'Greedy',
    'intervals': 'Intervals',
    'math': 'Math',
    'bit-manipulation': 'Bit Manip',
    'other': 'Other',
  }
  return map[cat] ?? cat
}

type TooltipPayload = { payload: { category: string; strength: number } }

function CustomTooltip({ active, payload }: { active?: boolean; payload?: TooltipPayload[] }) {
  if (!active || !payload?.length) return null
  const d = payload[0].payload
  return (
    <div className="rounded-lg border border-border bg-card px-3 py-2 text-xs shadow-lg">
      <p className="font-medium mb-0.5">{d.category}</p>
      <p className="text-muted-foreground">
        Strength: <span className="text-foreground font-semibold">{d.strength}%</span>
      </p>
    </div>
  )
}

const ALL_CATEGORIES = [
  'array', 'string', 'hash-map', 'two-pointers', 'sliding-window',
  'binary-search', 'stack', 'queue', 'linked-list', 'tree', 'trie',
  'heap', 'graph', 'advanced-graphs', 'dp', 'dp-2d', 'backtracking',
  'greedy', 'intervals', 'math', 'bit-manipulation', 'other',
]

export default function CategoryRadarChart({ stats }: Props) {
  const statsMap = new Map(stats.map((s) => [s.category, Math.round(s.strength ?? 0)]))

  const data = ALL_CATEGORIES.map((cat) => ({
    category: cat,
    label: shortLabel(cat),
    strength: statsMap.get(cat) ?? 0,
  }))

  return (
    <ResponsiveContainer width="100%" height={380}>
      <RadarChart data={data} margin={{ top: 10, right: 30, bottom: 10, left: 30 }}>
        <PolarGrid
          stroke="hsl(var(--border))"
          strokeOpacity={0.6}
        />
        <PolarAngleAxis
          dataKey="label"
          tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
          tickLine={false}
        />
        <Tooltip content={<CustomTooltip />} />
        <Radar
          dataKey="strength"
          stroke="hsl(var(--primary))"
          fill="hsl(var(--primary))"
          fillOpacity={0.15}
          strokeWidth={2}
          dot={{ r: 3, fill: 'hsl(var(--primary))', strokeWidth: 0 }}
        />
      </RadarChart>
    </ResponsiveContainer>
  )
}
