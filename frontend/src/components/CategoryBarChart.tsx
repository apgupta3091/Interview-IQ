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

type Props = { stats: CategoryStats[]; maxHeight?: number }

function barColor(s: number) {
  if (s >= 70) return '#10b981' // emerald-500
  if (s >= 40) return '#f59e0b' // amber-500
  return '#ef4444'              // red-500
}

type BarEntry = {
  category: string
  strength: number
  scoreReady: boolean
  problemCount: number
}

export default function CategoryBarChart({ stats, maxHeight }: Props) {
  const ready = [...stats]
    .filter((s) => s.score_ready)
    .sort((a, b) => (b.strength ?? 0) - (a.strength ?? 0))
    .map<BarEntry>((s) => ({
      category: s.category ?? '',
      strength: Math.round(s.strength ?? 0),
      scoreReady: true,
      problemCount: s.problem_count ?? 0,
    }))

  const pending = [...stats]
    .filter((s) => !s.score_ready)
    .sort((a, b) => (a.category ?? '').localeCompare(b.category ?? ''))
    .map<BarEntry>((s) => ({
      category: s.category ?? '',
      // Render pending bars with a placeholder value so they appear as thin stubs
      strength: 0,
      scoreReady: false,
      problemCount: s.problem_count ?? 0,
    }))

  const data = [...ready, ...pending]

  return (
    <div className="space-y-3">
      <ResponsiveContainer width="100%" height={maxHeight != null ? Math.min(data.length * 34 + 20, maxHeight) : data.length * 34 + 20}>
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
            tick={({ x, y, payload }: { x: string | number; y: string | number; payload: { value: string } }) => {
              const entry = data.find((d) => d.category === payload.value)
              const muted = !entry?.scoreReady
              return (
                <text
                  x={x}
                  y={y}
                  dy={4}
                  textAnchor="end"
                  fontSize={12}
                  fill={muted ? 'hsl(var(--muted-foreground))' : 'hsl(var(--foreground))'}
                >
                  {payload.value}
                </text>
              )
            }}
            axisLine={false}
            tickLine={false}
          />
          <Bar dataKey="strength" radius={[0, 4, 4, 0]} maxBarSize={14} isAnimationActive={false} style={{ cursor: 'default' }}>
            {data.map((entry) => (
              <Cell
                key={entry.category}
                fill={entry.scoreReady ? barColor(entry.strength) : 'hsl(var(--muted))'}
                fillOpacity={entry.scoreReady ? 1 : 0.6}
              />
            ))}
            <LabelList
              dataKey="strength"
              position="right"
              content={({ x, y, width, height, value, index }) => {
                if (index == null) return null
                const entry = data[index]
                if (!entry) return null
                const xPos = (x as number) + (width as number) + 6
                const yPos = (y as number) + (height as number) / 2 + 4
                if (entry.scoreReady) {
                  return (
                    <text x={xPos} y={yPos} fontSize={11} fill="hsl(var(--muted-foreground))">
                      {value}%
                    </text>
                  )
                }
                return (
                  <text x={xPos} y={yPos} fontSize={10} fill="hsl(var(--muted-foreground))">
                    {entry.problemCount}/3
                  </text>
                )
              }}
            />
          </Bar>
        </BarChart>
      </ResponsiveContainer>
      {pending.length > 0 && (
        <p className="text-xs text-muted-foreground px-1">
          Categories showing <span className="font-medium">N/3</span> need 3 submissions before a score is calculated.
        </p>
      )}
    </div>
  )
}
