import { X, ChevronDown } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { CATEGORIES } from '@/lib/constants'

const DIFFICULTIES = ['easy', 'medium', 'hard']

export const DATE_RANGE_OPTIONS = [
  { value: 'day',    label: 'Past day' },
  { value: 'week',   label: 'Past week' },
  { value: '2weeks', label: 'Past 2 weeks' },
  { value: 'month',  label: 'Past month' },
  { value: '3months',label: 'Past 3 months' },
] as const

export type DateRangeValue = typeof DATE_RANGE_OPTIONS[number]['value'] | ''

export const SCORE_RANGE_OPTIONS = [
  { value: '100-80', label: '100 – 80',  min: 80,  max: 100 },
  { value: '79-60',  label: '79 – 60',   min: 60,  max: 79  },
  { value: '59-40',  label: '59 – 40',   min: 40,  max: 59  },
  { value: '39-20',  label: '39 – 20',   min: 20,  max: 39  },
  { value: '19-0',   label: '19 – 0',    min: 0,   max: 19  },
] as const

export type ScoreRangeValue = typeof SCORE_RANGE_OPTIONS[number]['value'] | ''

type Props = {
  nameSearch: string
  onNameSearch: (v: string) => void
  dateRange: DateRangeValue
  onDateRangeChange: (v: DateRangeValue) => void
  selectedCategories: string[]
  onCategoriesChange: (v: string[]) => void
  selectedDifficulties: string[]
  onDifficultiesChange: (v: string[]) => void
  scoreRange: ScoreRangeValue
  onScoreRangeChange: (v: ScoreRangeValue) => void
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

export default function ProblemFilters({
  nameSearch, onNameSearch,
  dateRange, onDateRangeChange,
  selectedCategories, onCategoriesChange,
  selectedDifficulties, onDifficultiesChange,
  scoreRange, onScoreRangeChange,
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

        {/* Date range preset select */}
        <Select
          value={dateRange || '__all__'}
          onValueChange={(v) => onDateRangeChange(v === '__all__' ? '' : v as DateRangeValue)}
        >
          <SelectTrigger
            className={`h-8 w-auto min-w-[120px] text-xs font-normal gap-1${dateRange ? ' border-primary/50 bg-primary/5' : ''}`}
          >
            <SelectValue placeholder="Date range" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__all__" className="text-xs">All time</SelectItem>
            {DATE_RANGE_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value} className="text-xs">
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Score range preset select */}
        <Select
          value={scoreRange || '__all__'}
          onValueChange={(v) => onScoreRangeChange(v === '__all__' ? '' : v as ScoreRangeValue)}
        >
          <SelectTrigger
            className={`h-8 w-auto min-w-[110px] text-xs font-normal gap-1${scoreRange ? ' border-primary/50 bg-primary/5' : ''}`}
          >
            <SelectValue placeholder="Score" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__all__" className="text-xs">Any score</SelectItem>
            {SCORE_RANGE_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value} className="text-xs">
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

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
