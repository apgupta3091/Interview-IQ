import { X, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'

const CATEGORIES = [
  'array', 'string', 'hash-map', 'two-pointers', 'sliding-window',
  'binary-search', 'stack', 'queue', 'linked-list', 'tree', 'trie',
  'graph', 'advanced-graphs', 'heap', 'dp', 'dp-2d', 'backtracking',
  'greedy', 'intervals', 'math', 'bit-manipulation', 'other',
]

const DIFFICULTIES = ['easy', 'medium', 'hard']

type Props = {
  nameSearch: string
  onNameSearch: (v: string) => void
  dateFrom: string
  onDateFrom: (v: string) => void
  dateTo: string
  onDateTo: (v: string) => void
  selectedCategories: string[]
  onCategoriesChange: (v: string[]) => void
  selectedDifficulties: string[]
  onDifficultiesChange: (v: string[]) => void
  scoreMin: string
  onScoreMin: (v: string) => void
  scoreMax: string
  onScoreMax: (v: string) => void
  hasFilters: boolean
  onClear: () => void
}

function MultiSelect({
  label,
  options,
  selected,
  onChange,
}: {
  label: string
  options: string[]
  selected: string[]
  onChange: (v: string[]) => void
}) {
  const toggle = (val: string) =>
    onChange(selected.includes(val) ? selected.filter((s) => s !== val) : [...selected, val])

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="h-8 gap-1 text-xs font-normal">
          {label}
          {selected.length > 0 && (
            <span className="ml-0.5 rounded bg-primary/15 px-1 text-primary font-medium">
              {selected.length}
            </span>
          )}
          <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-52 p-2" align="start">
        <div className="space-y-1 max-h-56 overflow-y-auto pr-1">
          {options.map((opt) => (
            <div key={opt} className="flex items-center gap-2 px-1 py-0.5 rounded hover:bg-muted/50 cursor-pointer" onClick={() => toggle(opt)}>
              <Checkbox
                id={`${label}-${opt}`}
                checked={selected.includes(opt)}
                onCheckedChange={() => toggle(opt)}
              />
              <Label htmlFor={`${label}-${opt}`} className="text-xs cursor-pointer capitalize">
                {opt}
              </Label>
            </div>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}

export default function ProblemFilters({
  nameSearch, onNameSearch,
  dateFrom, onDateFrom,
  dateTo, onDateTo,
  selectedCategories, onCategoriesChange,
  selectedDifficulties, onDifficultiesChange,
  scoreMin, onScoreMin,
  scoreMax, onScoreMax,
  hasFilters, onClear,
}: Props) {
  return (
    <div className="space-y-2">
      {/* Row 1: name search */}
      <Input
        placeholder="Search by problem name…"
        value={nameSearch}
        onChange={(e) => onNameSearch(e.target.value)}
        className="h-8 text-sm"
        maxLength={200}
      />

      {/* Row 2: all filter controls */}
      <div className="flex flex-wrap items-center gap-2">
        <MultiSelect
          label="Category"
          options={CATEGORIES}
          selected={selectedCategories}
          onChange={onCategoriesChange}
        />
        <MultiSelect
          label="Difficulty"
          options={DIFFICULTIES}
          selected={selectedDifficulties}
          onChange={onDifficultiesChange}
        />

        <Separator orientation="vertical" className="h-5" />

        {/* Date range */}
        <div className="flex items-center gap-1">
          <span className="text-xs text-muted-foreground">From</span>
          <Input
            type="date"
            value={dateFrom}
            onChange={(e) => onDateFrom(e.target.value)}
            className="h-8 text-xs w-36"
          />
        </div>
        <div className="flex items-center gap-1">
          <span className="text-xs text-muted-foreground">To</span>
          <Input
            type="date"
            value={dateTo}
            onChange={(e) => onDateTo(e.target.value)}
            className="h-8 text-xs w-36"
          />
        </div>

        <Separator orientation="vertical" className="h-5" />

        {/* Score range */}
        <div className="flex items-center gap-1">
          <span className="text-xs text-muted-foreground">Score</span>
          <Input
            type="number"
            placeholder="min"
            value={scoreMin}
            onChange={(e) => onScoreMin(e.target.value)}
            className="h-8 text-xs w-16"
            min={0}
            max={100}
          />
          <span className="text-xs text-muted-foreground">–</span>
          <Input
            type="number"
            placeholder="max"
            value={scoreMax}
            onChange={(e) => onScoreMax(e.target.value)}
            className="h-8 text-xs w-16"
            min={0}
            max={100}
          />
        </div>

        {hasFilters && (
          <Button variant="ghost" size="sm" className="h-8 text-xs text-muted-foreground gap-1" onClick={onClear}>
            <X className="w-3.5 h-3.5" />
            Clear
          </Button>
        )}
      </div>
    </div>
  )
}
