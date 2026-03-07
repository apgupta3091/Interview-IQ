import {
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  Radar,
  ResponsiveContainer,
  Tooltip,
} from 'recharts'
import type { CategoryStats } from '@/types/api'

const ALL_CATEGORIES = [
  'array', 'string', 'hash-map', 'two-pointers', 'sliding-window',
  'binary-search', 'stack', 'queue', 'linked-list', 'tree', 'graph',
  'heap', 'dp', 'backtracking', 'greedy', 'math', 'other',
]

type Props = { stats: CategoryStats[] }

export default function SkillRadar({ stats }: Props) {
  // Build a map for O(1) lookup, then show all 17 categories (gaps are visible at 0)
  const strengthMap = new Map(stats.map((s) => [s.category, s.strength ?? 0]))

  const data = ALL_CATEGORIES.map((cat) => ({
    category: cat,
    strength: Math.round(strengthMap.get(cat) ?? 0),
  }))

  return (
    <ResponsiveContainer width="100%" height={380}>
      <RadarChart data={data} margin={{ top: 10, right: 30, bottom: 10, left: 30 }}>
        <PolarGrid />
        <PolarAngleAxis
          dataKey="category"
          tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
        />
        <Radar
          name="Strength"
          dataKey="strength"
          stroke="hsl(var(--primary))"
          fill="hsl(var(--primary))"
          fillOpacity={0.25}
        />
        <Tooltip
          formatter={(value) => [`${value}%`, 'Strength']}
          contentStyle={{
            fontSize: 12,
            borderRadius: 6,
            border: '1px solid hsl(var(--border))',
            background: 'hsl(var(--background))',
            color: 'hsl(var(--foreground))',
          }}
        />
      </RadarChart>
    </ResponsiveContainer>
  )
}
