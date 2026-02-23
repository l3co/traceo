import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Plus } from "lucide-react";
import { Link } from "react-router-dom";
import { api, type MissingResponse } from "@/shared/lib/api";
import { useAuth } from "@/shared/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import MissingCard from "@/features/missing/components/MissingCard";
import MissingDetailModal from "@/features/missing/components/MissingDetailModal";
import AdvancedSearchBar from "@/shared/components/AdvancedSearchBar";
import { useSearchFilter } from "@/shared/hooks/useSearchFilter";

const SEARCH_FIELDS: (keyof MissingResponse)[] = ["name", "nickname", "address"];

export default function MissingListPage() {
  const { t } = useTranslation();
  const { user } = useAuth();

  const [items, setItems] = useState<MissingResponse[]>([]);
  const [nextCursor, setNextCursor] = useState<string>();
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [selected, setSelected] = useState<MissingResponse | null>(null);

  const { filters, setFilter, reset, filtered, activeCount } = useSearchFilter(
    items,
    SEARCH_FIELDS,
  );

  const load = useCallback(async (cursor?: string) => {
    const isMore = !!cursor;
    if (isMore) setLoadingMore(true);
    else setLoading(true);

    try {
      const res = await api.listMissing(100, cursor);
      setItems((prev) => (isMore ? [...prev, ...res.items] : res.items));
      setNextCursor(res.next_cursor);
    } catch {
      // silent
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("missing.listTitle")}</h1>
        {user && (
          <Link to="/missing/new">
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              {t("missing.register")}
            </Button>
          </Link>
        )}
      </div>

      <AdvancedSearchBar
        filters={filters}
        setFilter={setFilter}
        reset={reset}
        activeCount={activeCount}
        resultCount={filtered.length}
        showStatus
      />

      {loading ? (
        <div className="flex justify-center py-16">
          <div className="flex flex-col items-center gap-3">
            <div className="h-8 w-8 rounded-full border-2 border-primary border-t-transparent animate-spin" />
            <p className="text-sm text-muted-foreground">{t("common.loading")}</p>
          </div>
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 gap-3">
          <p className="text-muted-foreground text-sm">{t("search.noResults")}</p>
          {activeCount > 0 && (
            <Button variant="outline" size="sm" onClick={reset}>
              {t("search.clearAll")}
            </Button>
          )}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filtered.map((item) => (
              <MissingCard
                key={item.id}
                item={item}
                onSelect={setSelected}
              />
            ))}
          </div>

          {nextCursor && activeCount === 0 && (
            <div className="flex justify-center">
              <Button
                variant="outline"
                onClick={() => load(nextCursor)}
                disabled={loadingMore}
              >
                {loadingMore ? t("common.loading") : t("missing.loadMore")}
              </Button>
            </div>
          )}
        </>
      )}

      {selected && (
        <MissingDetailModal
          item={selected}
          onClose={() => setSelected(null)}
        />
      )}
    </div>
  );
}
