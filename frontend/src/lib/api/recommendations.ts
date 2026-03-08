import type { RecommendationParams, RecommendationResult } from "@/types/api";
import client from "./client";

export const recommendations = {
  get: (params: RecommendationParams) => {
    const query = new URLSearchParams();
    if (params.categories) {
      params.categories.forEach((c) => query.append("category", c));
    }
    if (params.from) query.set("from", params.from);
    if (params.to) query.set("to", params.to);
    if (params.limit != null) query.set("limit", String(params.limit));

    return client
      .get<RecommendationResult>(`/api/recommendations?${query.toString()}`)
      .then((r) => r.data);
  },
};
