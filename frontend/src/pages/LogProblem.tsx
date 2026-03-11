import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import axios from 'axios'
import { Check, ChevronsUpDown, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { api } from '@/lib/api'
import type { ApiError, LeetCodeProblemSuggestion } from '@/types/api'
import { CATEGORIES } from '@/lib/constants'

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

  // Problem name + typeahead
  const [name, setName] = useState('')
  const [suggestions, setSuggestions] = useState<LeetCodeProblemSuggestion[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Form fields
  const [difficulty, setDifficulty] = useState('')
  const [selectedCategories, setSelectedCategories] = useState<string[]>([])
  const [categoryOpen, setCategoryOpen] = useState(false)
  const [attempts, setAttempts] = useState('1')
  const [lookedAtSolution, setLookedAtSolution] = useState(false)
  const [timeTaken, setTimeTaken] = useState('15')
  const [solutionType, setSolutionType] = useState<'none' | 'brute_force' | 'optimal'>('none')
  const [notes, setNotes] = useState('')

  // Debounced search as user types in the name field
  const handleNameChange = useCallback((value: string) => {
    setName(value)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    if (value.trim().length < 3) {
      setSuggestions([])
      setShowSuggestions(false)
      return
    }
    debounceRef.current = setTimeout(async () => {
      // 600 ms debounce keeps typeahead responsive while reducing request volume
      try {
        const results = await api.leetcodeProblems.search(value.trim())
        setSuggestions(results ?? [])
        setShowSuggestions(true)
      } catch {
        // silently ignore search errors — user can still type freely
      }
    }, 600)
  }, [])

  // When a suggestion is selected, auto-fill name, difficulty, and categories
  function selectSuggestion(s: LeetCodeProblemSuggestion) {
    setName(s.title ?? '')
    if (s.difficulty) setDifficulty(s.difficulty)
    if (s.tags && s.tags.length > 0) {
      // Only pre-select tags that are in our known category list
      const valid = s.tags.filter((t) => (CATEGORIES as readonly string[]).includes(t))
      if (valid.length > 0) setSelectedCategories(valid)
    }
    setShowSuggestions(false)
    setSuggestions([])
  }

  // Close suggestion dropdown on outside click
  useEffect(() => {
    function onBlur() { setShowSuggestions(false) }
    document.addEventListener('mousedown', onBlur)
    return () => document.removeEventListener('mousedown', onBlur)
  }, [])

  function toggleCategory(cat: string) {
    setSelectedCategories((prev) =>
      prev.includes(cat) ? prev.filter((c) => c !== cat) : [...prev, cat],
    )
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (selectedCategories.length === 0) { toast.error('Please select at least one category'); return }
    if (!difficulty) { toast.error('Please select a difficulty'); return }
    setLoading(true)
    try {
      await api.problems.log({
        name,
        categories: selectedCategories,
        difficulty,
        attempts: parseInt(attempts) || 1,
        looked_at_solution: lookedAtSolution,
        time_taken_mins: parseInt(timeTaken) || 1,
        solution_type: solutionType,
        ...(notes.trim() ? { notes: notes.trim() } : {}),
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

            {/* Problem name with typeahead */}
            <div className="space-y-1.5 relative">
              <Label htmlFor="name">Problem name</Label>
              <Input
                id="name"
                placeholder="e.g. Two Sum"
                required
                autoComplete="off"
                maxLength={200}
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
                onFocus={() => suggestions.length > 0 && setShowSuggestions(true)}
              />
              {showSuggestions && suggestions.length > 0 && (
                <div
                  className="absolute z-50 w-full top-full mt-1 rounded-md border border-border bg-popover shadow-md overflow-hidden"
                  onMouseDown={(e) => { e.preventDefault(); e.stopPropagation() }} // prevent blur and stop document listener
                >
                  <ul className="max-h-52 overflow-y-auto py-1">
                    {suggestions.map((s) => (
                      <li
                        key={s.lc_id}
                        className="flex items-center justify-between px-3 py-1.5 text-sm cursor-pointer hover:bg-accent"
                        onClick={() => selectSuggestion(s)}
                      >
                        <span className="truncate">{s.title}</span>
                        <span className={cn(
                          'ml-2 shrink-0 text-xs font-medium',
                          s.difficulty === 'easy' && 'text-emerald-500',
                          s.difficulty === 'medium' && 'text-amber-500',
                          s.difficulty === 'hard' && 'text-red-500',
                        )}>
                          {s.difficulty}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>

            {/* Multi-select categories */}
            <div className="space-y-1.5">
              <Label>Categories</Label>
              {selectedCategories.length > 0 && (
                <div className="flex flex-wrap gap-1.5 mb-2">
                  {selectedCategories.map((cat) => (
                    <Badge
                      key={cat}
                      variant="secondary"
                      className="gap-1 pl-2 pr-1 text-xs cursor-pointer"
                      onClick={() => toggleCategory(cat)}
                    >
                      {cat}
                      <X className="h-3 w-3" />
                    </Badge>
                  ))}
                </div>
              )}
              <Popover open={categoryOpen} onOpenChange={setCategoryOpen}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    role="combobox"
                    aria-expanded={categoryOpen}
                    className="w-full justify-between font-normal"
                  >
                    {selectedCategories.length === 0
                      ? 'Select categories…'
                      : `${selectedCategories.length} selected`}
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
                            onSelect={() => toggleCategory(c)}
                          >
                            <Check className={cn('mr-2 h-4 w-4', selectedCategories.includes(c) ? 'opacity-100' : 'opacity-0')} />
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
              <Select value={difficulty} onValueChange={setDifficulty}>
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

            {/* Solution type */}
            <div className="space-y-2">
              <Label>Solution achieved</Label>
              <p className="text-xs text-muted-foreground -mt-1">Brute-force reduces score by 15 pts; optimal carries no penalty.</p>
              <div className="grid grid-cols-3 gap-2">
                {(
                  [
                    { value: 'none',        label: 'Not specified', desc: 'No impact' },
                    { value: 'brute_force', label: 'Brute force',   desc: '−15 pts' },
                    { value: 'optimal',     label: 'Optimal',       desc: 'No impact' },
                  ] as const
                ).map(({ value, label, desc }) => (
                  <button
                    key={value}
                    type="button"
                    onClick={() => setSolutionType(value)}
                    className={cn(
                      'flex flex-col items-center justify-center rounded-lg border px-2 py-2.5 text-sm transition-colors',
                      solutionType === value
                        ? value === 'brute_force'
                          ? 'border-amber-500 bg-amber-500/10 text-amber-600 dark:text-amber-400'
                          : value === 'optimal'
                          ? 'border-emerald-500 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                          : 'border-primary bg-primary/10 text-primary'
                        : 'border-border/60 bg-muted/20 text-muted-foreground hover:bg-muted/40',
                    )}
                  >
                    <span className="font-medium leading-tight">{label}</span>
                    <span className="text-xs opacity-70 mt-0.5">{desc}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Notes */}
            <div className="space-y-1.5">
              <Label htmlFor="notes">Notes <span className="text-muted-foreground font-normal">(optional)</span></Label>
              <Textarea
                id="notes"
                placeholder="Approach, edge cases, things to remember…"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
                className="resize-none"
              />
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
