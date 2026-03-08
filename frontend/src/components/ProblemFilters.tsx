import { useState } from 'react'
import { X, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { CATEGORIES } from '@/lib/constants'

const DIFFICULTIES = ['easy', 'medium', 'hard']

type Props = {
  nameSearch: string
  onNameSearch: (v: string) => void
  dateFrom: string
  dateTo: string
  onApplyDateRange: (from: string, to: string) => void
  selectedCategories: string[]
  onCategoriesChange: (v: string[]) => void
  selectedDifficulties: string[]
  onDifficultiesChange: (v: string[]) => void
  scoreMin: string
  scoreMax: string
  onApplyScoreRange: (min: string, max: string) => void
  /** True when any filter is active — shows the Clear button. */
  hasFilters: boolean
  onClear: () => void
}

export function MultiSelect({
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
            <div
              key={opt}
              className="flex items-center gap-2 px-1 py-0.5 rounded hover:bg-muted/50 cursor-pointer"
              onClick={() => toggle(opt)}
            >
              <Checkbox
                id={`${label}-${opt}`}
                checked={selected.includes(opt)}
                onClick={(e) => e.stopPropagation()}
                onCheckedChange={() => toggle(opt)}
              />
              <Label
                htmlFor={`${label}-${opt}`}
                className="text-xs cursor-pointer capitalize"
                onClick={(e) => e.stopPropagation()}
              >
                {opt}
              </Label>
            </div>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}

function DateRangeFilter({
  dateFrom,
  dateTo,
  onApply,
}: {
  dateFrom: string
  dateTo: string
  onApply: (from: string, to: string) => void
}) {
  const [open, setOpen] = useState(false)
  const [localFrom, setLocalFrom] = useState(dateFrom)
  const [localTo, setLocalTo] = useState(dateTo)

  function handleOpenChange(o: boolean) {
    if (o) {
      setLocalFrom(dateFrom)
      setLocalTo(dateTo)
    }
    setOpen(o)
  }

  function handleApply() {
    onApply(localFrom, localTo)
    setOpen(false)
  }

  const isActive = !!(dateFrom || dateTo)
  let buttonLabel = 'Date range'
  if (dateFrom && dateTo) buttonLabel = `${dateFrom} – ${dateTo}`
  else if (dateFrom) buttonLabel = `From ${dateFrom}`
  else if (dateTo) buttonLabel = `To ${dateTo}`

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={`h-8 gap-1 text-xs font-normal${isActive ? ' border-primary/50 bg-primary/5' : ''}`}
        >
          {buttonLabel}
          <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-60 p-3" align="start">
        <div className="space-y-3">
          <div className="space-y-1.5">
            <Label className="text-xs">From</Label>
            <Input
              type="date"
              value={localFrom}
              onChange={(e) => setLocalFrom(e.target.value)}
              className="h-8 text-xs"
            />
          </div>
          <div className="space-y-1.5">
            <Label className="text-xs">To</Label>
            <Input
              type="date"
              value={localTo}
              onChange={(e) => setLocalTo(e.target.value)}
              className="h-8 text-xs"
            />
          </div>
          <Button size="sm" className="w-full h-8 text-xs" onClick={handleApply}>
            Apply
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}

function ScoreRangeFilter({
  scoreMin,
  scoreMax,
  onApply,
}: {
  scoreMin: string
  scoreMax: string
  onApply: (min: string, max: string) => void
}) {
  const [open, setOpen] = useState(false)
  const [localMin, setLocalMin] = useState(scoreMin)
  const [localMax, setLocalMax] = useState(scoreMax)

  function handleOpenChange(o: boolean) {
    if (o) {
      setLocalMin(scoreMin)
      setLocalMax(scoreMax)
    }
    setOpen(o)
  }

  function handleApply() {
    onApply(localMin, localMax)
    setOpen(false)
  }

  const isActive = !!(scoreMin || scoreMax)
  let buttonLabel = 'Score'
  if (scoreMin && scoreMax) buttonLabel = `Score ${scoreMin}–${scoreMax}`
  else if (scoreMin) buttonLabel = `Score ≥ ${scoreMin}`
  else if (scoreMax) buttonLabel = `Score ≤ ${scoreMax}`

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={`h-8 gap-1 text-xs font-normal${isActive ? ' border-primary/50 bg-primary/5' : ''}`}
        >
          {buttonLabel}
          <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-48 p-3" align="start">
        <div className="space-y-3">
          <div className="space-y-1.5">
            <Label className="text-xs">Min score</Label>
            <Input
              type="number"
              placeholder="0"
              value={localMin}
              onChange={(e) => setLocalMin(e.target.value)}
              className="h-8 text-xs"
              min={0}
              max={100}
            />
          </div>
          <div className="space-y-1.5">
            <Label className="text-xs">Max score</Label>
            <Input
              type="number"
              placeholder="100"
              value={localMax}
              onChange={(e) => setLocalMax(e.target.value)}
              className="h-8 text-xs"
              min={0}
              max={100}
            />
          </div>
          <Button size="sm" className="w-full h-8 text-xs" onClick={handleApply}>
            Apply
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}

export default function ProblemFilters({
  nameSearch, onNameSearch,
  dateFrom, dateTo, onApplyDateRange,
  selectedCategories, onCategoriesChange,
  selectedDifficulties, onDifficultiesChange,
  scoreMin, scoreMax, onApplyScoreRange,
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

      {/* Row 2: filter controls */}
      <div className="flex flex-wrap items-center gap-2">
        <MultiSelect
          label="Category"
          options={[...CATEGORIES]}
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

        <DateRangeFilter dateFrom={dateFrom} dateTo={dateTo} onApply={onApplyDateRange} />
        <ScoreRangeFilter scoreMin={scoreMin} scoreMax={scoreMax} onApply={onApplyScoreRange} />

        {hasFilters && (
          <>
            <Separator orientation="vertical" className="h-5" />
            <Button variant="ghost" size="sm" className="h-8 text-xs text-muted-foreground gap-1" onClick={onClear}>
              <X className="w-3.5 h-3.5" />
              Clear
            </Button>
          </>
        )}
      </div>
    </div>
  )
}
