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

const DIFFICULTIES = ['easy', 'medium', 'hard']

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
      await api.problems.log({
        name,
        category,
        difficulty,
        attempts,
        looked_at_solution: lookedAtSolution,
        time_taken_mins: timeTaken,
      })
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
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Log a problem</CardTitle>
          <CardDescription>Record a problem you solved during practice.</CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
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

            <div className="space-y-1.5">
              <Label>Category</Label>
              <Select onValueChange={setCategory}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a category" />
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
                  <SelectValue placeholder="Select difficulty" />
                </SelectTrigger>
                <SelectContent>
                  {DIFFICULTIES.map((d) => (
                    <SelectItem key={d} value={d}>{d}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

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
              <Label htmlFor="time">Time taken (minutes)</Label>
              <Input
                id="time"
                type="number"
                min={1}
                value={timeTaken}
                onChange={(e) => setTimeTaken(Number(e.target.value))}
              />
            </div>

            <div className="flex items-center gap-2">
              <Checkbox
                id="looked"
                checked={lookedAtSolution}
                onCheckedChange={(v) => setLookedAtSolution(v === true)}
              />
              <Label htmlFor="looked">I looked at the solution</Label>
            </div>
          </CardContent>

          <CardFooter>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? 'Logging…' : 'Log problem'}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}
