import type { RecommendationParams, RecommendationResult } from "@/types/api";
import client from "./client";

export const recommendations = {
  get: (params: RecommendationParams) => {
    const query = new URLSearchParams();
    if (params.category) query.set("category", params.category);
    if (params.limit != null) query.set("limit", String(params.limit));

    return client
      .get<RecommendationResult>(`/api/recommendations?${query.toString()}`)
      .then((r) => r.data);
  },
};
