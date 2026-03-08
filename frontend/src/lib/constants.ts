export const CATEGORIES = [
  'array', 'string', 'hash-map', 'two-pointers', 'sliding-window',
  'binary-search', 'stack', 'queue', 'linked-list', 'tree', 'trie',
  'graph', 'advanced-graphs', 'heap', 'dp', 'dp-2d', 'backtracking',
  'greedy', 'intervals', 'math', 'bit-manipulation', 'other',
] as const

export type Category = typeof CATEGORIES[number]
