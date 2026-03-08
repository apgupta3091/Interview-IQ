import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Cell,
  LabelList,
  ResponsiveContainer,
} from 'recharts'
import type { CategoryStats } from '@/types/api'

type Props = { stats: CategoryStats[] }

function barColor(s: number) {
  if (s >= 70) return '#10b981' // emerald-500
  if (s >= 40) return '#f59e0b' // amber-500
  return '#ef4444'              // red-500
}

export default function CategoryBarChart({ stats }: Props) {
  const data = [...stats]
    .sort((a, b) => (b.strength ?? 0) - (a.strength ?? 0))
    .map((s) => ({ category: s.category, strength: Math.round(s.strength ?? 0) }))

  return (
    <ResponsiveContainer width="100%" height={data.length * 34 + 20}>
      <BarChart
        data={data}
        layout="vertical"
        margin={{ top: 0, right: 48, bottom: 0, left: 8 }}
        barCategoryGap="30%"
        style={{ cursor: 'default' }}
      >
        <XAxis
          type="number"
          domain={[0, 100]}
          tickCount={6}
          tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
          axisLine={false}
          tickLine={false}
        />
        <YAxis
          type="category"
          dataKey="category"
          width={90}
          tick={{ fontSize: 12, fill: 'hsl(var(--foreground))' }}
          axisLine={false}
          tickLine={false}
        />
        <Bar dataKey="strength" radius={[0, 4, 4, 0]} maxBarSize={14} style={{ cursor: 'default' }}>
          {data.map((entry) => (
            <Cell key={entry.category} fill={barColor(entry.strength)} />
          ))}
          <LabelList
            dataKey="strength"
            position="right"
            formatter={(v: string | number) => `${v}%`}
            style={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
          />
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  )
}
