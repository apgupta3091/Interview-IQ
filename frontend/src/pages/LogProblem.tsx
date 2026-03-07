import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Check, ChevronsUpDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { cn } from '@/lib/utils'
import { api } from '@/lib/api'
import type { ApiError } from '@/types/api'

const CATEGORIES = [
  'array', 'string', 'hash-map', 'two-pointers', 'sliding-window',
  'binary-search', 'stack', 'linked-list', 'tree', 'trie', 'heap',
  'graph', 'advanced-graphs', 'dp', 'dp-2d', 'backtracking',
  'greedy', 'intervals', 'math', 'bit-manipulation', 'queue', 'other',
]

const DIFFICULTIES = [
  { value: 'easy',   label: 'Easy',   color: 'text-emerald-500' },
  { value: 'medium', label: 'Medium', color: 'text-amber-500' },
  { value: 'hard',   label: 'Hard',   color: 'text-red-500' },
]

function NumberInput({
  id, label, value, onChange, min = 1,
}: {
  id: string
  label: string
  value: string
  onChange: (v: string) => void
  min?: number
}) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        type="number"
        min={min}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onBlur={(e) => {
          const n = parseInt(e.target.value)
          onChange(isNaN(n) || n < min ? String(min) : String(n))
        }}
      />
    </div>
  )
}

export default function LogProblem() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)

  const [name, setName] = useState('')
  const [category, setCategory] = useState('')
  const [categoryOpen, setCategoryOpen] = useState(false)
  const [difficulty, setDifficulty] = useState('')
  const [attempts, setAttempts] = useState('1')
  const [lookedAtSolution, setLookedAtSolution] = useState(false)
  const [timeTaken, setTimeTaken] = useState('15')

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
        attempts: parseInt(attempts) || 1,
        looked_at_solution: lookedAtSolution,
        time_taken_mins: parseInt(timeTaken) || 1,
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
    <div className="max-w-lg mx-auto animate-fade-up">
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Log a problem</h1>
        <p className="text-sm text-muted-foreground mt-1">Record a problem you solved during practice.</p>
      </div>

      <Card className="border-border/60 shadow-sm">
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-5 pt-6">

            {/* Problem name */}
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

            {/* Category combobox */}
            <div className="space-y-1.5">
              <Label>Category</Label>
              <Popover open={categoryOpen} onOpenChange={setCategoryOpen}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={categoryOpen}
                    className="w-full justify-between font-normal"
                  >
                    {category || 'Select a category…'}
                    <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-full p-0" align="start">
                  <Command>
                    <CommandInput placeholder="Search categories…" />
                    <CommandList style={{ maxHeight: '220px', overflowY: 'auto' }}>
                      <CommandEmpty>No category found.</CommandEmpty>
                      <CommandGroup>
                        {CATEGORIES.map((c) => (
                          <CommandItem
                            key={c}
                            value={c}
                            onSelect={(val) => {
                              setCategory(val)
                              setCategoryOpen(false)
                            }}
                          >
                            <Check className={cn('mr-2 h-4 w-4', category === c ? 'opacity-100' : 'opacity-0')} />
                            {c}
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </CommandList>
                  </Command>
                </PopoverContent>
              </Popover>
            </div>

            {/* Difficulty */}
            <div className="space-y-1.5">
              <Label>Difficulty</Label>
              <Select onValueChange={setDifficulty}>
                <SelectTrigger>
                  <SelectValue placeholder="Select difficulty…" />
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

            {/* Attempts + Time */}
            <div className="grid grid-cols-2 gap-4">
              <NumberInput id="attempts" label="Attempts" value={attempts} onChange={setAttempts} min={1} />
              <NumberInput id="time" label="Time (minutes)" value={timeTaken} onChange={setTimeTaken} min={1} />
            </div>

            {/* Looked at solution */}
            <div className="flex items-center gap-2.5 rounded-lg border border-border/60 bg-muted/20 px-3 py-2.5">
              <Checkbox
                id="looked"
                checked={lookedAtSolution}
                onCheckedChange={(v) => setLookedAtSolution(v === true)}
              />
              <div>
                <Label htmlFor="looked" className="cursor-pointer text-sm font-medium">Looked at the solution</Label>
                <p className="text-xs text-muted-foreground">Reduces your score by 25 points</p>
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
