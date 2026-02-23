import { useMemo, useState } from "react";

export interface SearchFilters {
  query: string;
  gender: string;
  eyes: string;
  hair: string;
  skin: string;
  status: string;
  ageMin: string;
  ageMax: string;
}

export const EMPTY_FILTERS: SearchFilters = {
  query: "",
  gender: "",
  eyes: "",
  hair: "",
  skin: "",
  status: "",
  ageMin: "",
  ageMax: "",
};

export function useSearchFilter<T extends object>(
  items: T[],
  searchFields: (keyof T)[],
) {
  const [filters, setFilters] = useState<SearchFilters>(EMPTY_FILTERS);

  const activeCount = useMemo(
    () =>
      Object.entries(filters).filter(([, v]) => v !== "").length,
    [filters],
  );

  const filtered = useMemo(() => {
    return items.filter((item) => {
      const r = item as Record<string, unknown>;

      if (filters.query) {
        const q = filters.query.toLowerCase();
        const matches = searchFields.some((field) => {
          const val = r[field as string];
          return typeof val === "string" && val.toLowerCase().includes(q);
        });
        if (!matches) return false;
      }

      if (filters.gender && r.gender !== filters.gender) return false;
      if (filters.eyes && r.eyes !== filters.eyes) return false;
      if (filters.hair && r.hair !== filters.hair) return false;
      if (filters.skin && r.skin !== filters.skin) return false;
      if (filters.status && r.status !== filters.status) return false;

      const age = typeof r.age === "number" ? r.age : 0;
      if (filters.ageMin && age < parseInt(filters.ageMin)) return false;
      if (filters.ageMax && age > parseInt(filters.ageMax)) return false;

      return true;
    });
  }, [items, filters, searchFields]);

  const setFilter = <K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) =>
    setFilters((prev) => ({ ...prev, [key]: value }));

  const reset = () => setFilters(EMPTY_FILTERS);

  return { filters, setFilter, reset, filtered, activeCount };
}
