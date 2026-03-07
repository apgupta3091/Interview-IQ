import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { api } from '@/lib/api'
import type { ApiError } from '@/types/api'

const CATEGORIES = [
  'array', 'string', 'hash-map', 'two-pointers', 'sliding-window',
  'binary-search', 'stack', 'queue', 'linked-list', 'tree', 'graph',
  'heap', 'dp', 'backtracking', 'greedy', 'math', 'other',
]

const DIFFICULTIES = [
  { value: 'easy',   label: 'Easy',   color: 'text-emerald-500' },
  { value: 'medium', label: 'Medium', color: 'text-amber-500' },
  { value: 'hard',   label: 'Hard',   color: 'text-red-500' },
]

export default function LogProblem() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [name, setName] = useState('')
  const [category, setCategory] = useState('')
  const [difficulty, setDifficulty] = useState('')
  const [attempts, setAttempts] = useState(1)
  const [lookedAtSolution, setLookedAtSolution] = useState(false)
  const [timeTaken, setTimeTaken] = useState(15)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!category) { toast.error('Please select a category'); return }
    if (!difficulty) { toast.error('Please select a difficulty'); return }
    setLoading(true)
    try {
      await api.problems.log({ name, category, difficulty, attempts, looked_at_solution: lookedAtSolution, time_taken_mins: timeTaken })
      toast.success('Problem logged!')
      navigate('/problems')
    } catch (err) {
      if (axios.isAxiosError(err)) {
        const msg = (err.response?.data as ApiError)?.error ?? 'Failed to log problem'
        toast.error(msg)
      } else {
        toast.error('Unexpected error')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-lg mx-auto animate-fade-up">
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Log a problem</h1>
        <p className="text-sm text-muted-foreground mt-1">Record a problem you solved during practice.</p>
      </div>

      <Card className="border-border/60 shadow-sm">
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-5 pt-6">
            <div className="space-y-1.5">
              <Label htmlFor="name">Problem name</Label>
              <Input
                id="name"
                placeholder="e.g. Two Sum"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>Category</Label>
                <Select onValueChange={setCategory}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select…" />
                  </SelectTrigger>
                  <SelectContent>
                    {CATEGORIES.map((c) => (
                      <SelectItem key={c} value={c}>{c}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-1.5">
                <Label>Difficulty</Label>
                <Select onValueChange={setDifficulty}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select…" />
                  </SelectTrigger>
                  <SelectContent>
                    {DIFFICULTIES.map((d) => (
                      <SelectItem key={d.value} value={d.value}>
                        <span className={d.color}>{d.label}</span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="attempts">Attempts</Label>
                <Input
                  id="attempts"
                  type="number"
                  min={1}
                  value={attempts}
                  onChange={(e) => setAttempts(Number(e.target.value))}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="time">Time (minutes)</Label>
                <Input
                  id="time"
                  type="number"
                  min={1}
                  value={timeTaken}
                  onChange={(e) => setTimeTaken(Number(e.target.value))}
                />
              </div>
            </div>

            <div className="flex items-center gap-2.5 rounded-lg border border-border/60 bg-muted/20 px-3 py-2.5">
              <Checkbox
                id="looked"
                checked={lookedAtSolution}
                onCheckedChange={(v) => setLookedAtSolution(v === true)}
              />
              <div>
                <Label htmlFor="looked" className="cursor-pointer text-sm font-medium">Looked at the solution</Label>
                <p className="text-xs text-muted-foreground">This will reduce your score by 25 points</p>
              </div>
            </div>
          </CardContent>

          <CardFooter className="pt-2">
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? 'Logging…' : 'Log problem'}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}
