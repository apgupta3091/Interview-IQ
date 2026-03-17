import type { ProblemNote, NoteListResponse } from '@/types/api'
import client from './client'

export const notes = {
  list: (problemID: number): Promise<NoteListResponse> =>
    client.get<NoteListResponse>(`/api/problems/${problemID}/notes`).then((r) => r.data),

  create: (problemID: number, content: string): Promise<ProblemNote> =>
    client.post<ProblemNote>(`/api/problems/${problemID}/notes`, { content }).then((r) => r.data),

  update: (noteID: number, content: string): Promise<ProblemNote> =>
    client.put<ProblemNote>(`/api/notes/${noteID}`, { content }).then((r) => r.data),

  delete: (noteID: number): Promise<void> =>
    client.delete(`/api/notes/${noteID}`).then(() => undefined),
}
