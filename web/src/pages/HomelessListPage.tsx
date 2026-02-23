import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { Plus } from "lucide-react";
import { api, type HomelessResponse } from "@/shared/lib/api";
import { useAuth } from "@/shared/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import HomelessCard from "@/features/homeless/components/HomelessCard";
import AdvancedSearchBar from "@/shared/components/AdvancedSearchBar";
import { useSearchFilter } from "@/shared/hooks/useSearchFilter";

const SEARCH_FIELDS: (keyof HomelessResponse)[] = ["name", "nickname", "address"];

export default function HomelessListPage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [items, setItems] = useState<HomelessResponse[]>([]);
  const [loading, setLoading] = useState(true);

  const { filters, setFilter, reset, filtered, activeCount } = useSearchFilter(
    items,
    SEARCH_FIELDS,
  );

  useEffect(() => {
    api
      .listHomeless()
      .then(setItems)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t("homeless.title")}</h1>
        {user && (
          <Link to="/homeless/new">
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              {t("homeless.register")}
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
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {filtered.map((item) => (
            <HomelessCard key={item.id} item={item} />
          ))}
        </div>
      )}
    </div>
  );
}
