import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Search,
  SlidersHorizontal,
  X,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import type { SearchFilters } from "@/shared/hooks/useSearchFilter";
import { EMPTY_FILTERS } from "@/shared/hooks/useSearchFilter";
import {
  GENDER_OPTIONS,
  EYE_OPTIONS,
  HAIR_OPTIONS,
  SKIN_OPTIONS,
  STATUS_OPTIONS,
  getLabel,
} from "@/features/missing/constants";

interface Props {
  filters: SearchFilters;
  setFilter: <K extends keyof SearchFilters>(key: K, value: SearchFilters[K]) => void;
  reset: () => void;
  activeCount: number;
  resultCount: number;
  showStatus?: boolean;
}

interface SelectChipProps {
  label: string;
  value: string;
  options: { value: string; labelPt: string; labelEn: string }[];
  lang: string;
  onChange: (v: string) => void;
}

function SelectChip({ label, value, options, lang, onChange }: SelectChipProps) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
        {label}
      </span>
      <div className="flex flex-wrap gap-1.5">
        {options.map((opt) => {
          const active = value === opt.value;
          return (
            <button
              key={opt.value}
              type="button"
              onClick={() => onChange(active ? "" : opt.value)}
              className={`px-3 py-1 rounded-full text-xs font-medium border transition-all duration-150 ${
                active
                  ? "bg-primary text-primary-foreground border-primary shadow-sm"
                  : "bg-background text-muted-foreground border-border hover:border-primary/50 hover:text-foreground"
              }`}
            >
              {getLabel(options, opt.value, lang)}
            </button>
          );
        })}
      </div>
    </div>
  );
}

export default function AdvancedSearchBar({
  filters,
  setFilter,
  reset,
  activeCount,
  resultCount,
  showStatus = false,
}: Props) {
  const { t, i18n } = useTranslation();
  const lang = i18n.language;
  const [open, setOpen] = useState(false);

  const activeChips: { key: keyof SearchFilters; label: string }[] = [];

  if (filters.gender)
    activeChips.push({ key: "gender", label: getLabel(GENDER_OPTIONS, filters.gender, lang) });
  if (filters.eyes)
    activeChips.push({ key: "eyes", label: getLabel(EYE_OPTIONS, filters.eyes, lang) });
  if (filters.hair)
    activeChips.push({ key: "hair", label: getLabel(HAIR_OPTIONS, filters.hair, lang) });
  if (filters.skin)
    activeChips.push({ key: "skin", label: getLabel(SKIN_OPTIONS, filters.skin, lang) });
  if (filters.status)
    activeChips.push({ key: "status", label: getLabel(STATUS_OPTIONS, filters.status, lang) });
  if (filters.ageMin)
    activeChips.push({ key: "ageMin", label: `${t("search.ageMin")}: ${filters.ageMin}` });
  if (filters.ageMax)
    activeChips.push({ key: "ageMax", label: `${t("search.ageMax")}: ${filters.ageMax}` });

  return (
    <div className="rounded-2xl border bg-card shadow-sm overflow-hidden">
      {/* Search bar row */}
      <div className="flex items-center gap-3 px-4 py-3">
        <Search className="h-4 w-4 text-muted-foreground shrink-0" />
        <Input
          value={filters.query}
          onChange={(e) => setFilter("query", e.target.value)}
          placeholder={t("search.placeholder")}
          className="border-0 shadow-none focus-visible:ring-0 px-0 text-base bg-transparent"
        />
        {activeCount > 0 && (
          <Badge variant="secondary" className="shrink-0 rounded-full tabular-nums">
            {activeCount}
          </Badge>
        )}
        <div className="flex items-center gap-1 shrink-0">
          {activeCount > 0 && (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-muted-foreground hover:text-foreground"
              onClick={reset}
              title={t("search.clearAll")}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
          <Button
            type="button"
            variant={open ? "secondary" : "ghost"}
            size="sm"
            className="gap-1.5 rounded-full"
            onClick={() => setOpen((v) => !v)}
          >
            <SlidersHorizontal className="h-3.5 w-3.5" />
            <span className="text-xs">{t("search.filters")}</span>
            {open ? (
              <ChevronUp className="h-3 w-3" />
            ) : (
              <ChevronDown className="h-3 w-3" />
            )}
          </Button>
        </div>
      </div>

      {/* Active filter chips */}
      {activeChips.length > 0 && (
        <div className="flex flex-wrap gap-1.5 px-4 pb-3">
          {activeChips.map(({ key, label }) => (
            <span
              key={key}
              className="inline-flex items-center gap-1 bg-primary/10 text-primary text-xs font-medium px-2.5 py-0.5 rounded-full"
            >
              {label}
              <button
                type="button"
                onClick={() => setFilter(key, EMPTY_FILTERS[key])}
                className="hover:text-primary/70 transition-colors"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      )}

      {/* Expandable filter panel */}
      {open && (
        <div className="border-t bg-muted/30 px-4 py-5 space-y-5">
          <SelectChip
            label={t("missing.gender")}
            value={filters.gender}
            options={GENDER_OPTIONS}
            lang={lang}
            onChange={(v) => setFilter("gender", v)}
          />
          <SelectChip
            label={t("missing.eyes")}
            value={filters.eyes}
            options={EYE_OPTIONS}
            lang={lang}
            onChange={(v) => setFilter("eyes", v)}
          />
          <SelectChip
            label={t("missing.hair")}
            value={filters.hair}
            options={HAIR_OPTIONS}
            lang={lang}
            onChange={(v) => setFilter("hair", v)}
          />
          <SelectChip
            label={t("missing.skin")}
            value={filters.skin}
            options={SKIN_OPTIONS}
            lang={lang}
            onChange={(v) => setFilter("skin", v)}
          />
          {showStatus && (
            <SelectChip
              label={t("missing.status")}
              value={filters.status}
              options={STATUS_OPTIONS}
              lang={lang}
              onChange={(v) => setFilter("status", v)}
            />
          )}

          {/* Age range */}
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              {t("search.ageRange")}
            </span>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                min={0}
                max={120}
                placeholder={t("search.ageMin")}
                value={filters.ageMin}
                onChange={(e) => setFilter("ageMin", e.target.value)}
                className="w-24 h-8 text-sm"
              />
              <span className="text-muted-foreground text-sm">—</span>
              <Input
                type="number"
                min={0}
                max={120}
                placeholder={t("search.ageMax")}
                value={filters.ageMax}
                onChange={(e) => setFilter("ageMax", e.target.value)}
                className="w-24 h-8 text-sm"
              />
              <span className="text-xs text-muted-foreground">{t("search.years")}</span>
            </div>
          </div>

          {/* Result count + clear */}
          <div className="flex items-center justify-between pt-1 border-t">
            <p className="text-sm text-muted-foreground">
              <span className="font-semibold text-foreground">{resultCount}</span>{" "}
              {t("search.results")}
            </p>
            {activeCount > 0 && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={reset}
                className="text-xs text-muted-foreground hover:text-foreground"
              >
                {t("search.clearAll")}
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
