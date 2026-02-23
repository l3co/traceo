import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { Search, Heart } from "lucide-react";
import { api, type MissingResponse, type HomelessResponse } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import MissingCard from "@/features/missing/components/MissingCard";
import MissingDetailModal from "@/features/missing/components/MissingDetailModal";
import HomelessCard from "@/features/homeless/components/HomelessCard";
import Head from "@/shared/components/Head";

export default function HomePage() {
  const { t } = useTranslation();

  const [missingItems, setMissingItems] = useState<MissingResponse[]>([]);
  const [homelessItems, setHomelessItems] = useState<HomelessResponse[]>([]);
  const [loadingMissing, setLoadingMissing] = useState(true);
  const [loadingHomeless, setLoadingHomeless] = useState(true);
  const [nextCursor, setNextCursor] = useState<string>();
  const [loadingMore, setLoadingMore] = useState(false);
  const [selected, setSelected] = useState<MissingResponse | null>(null);

  const loadMissing = useCallback(async (cursor?: string) => {
    const isMore = !!cursor;
    if (isMore) setLoadingMore(true);
    else setLoadingMissing(true);

    try {
      const res = await api.listMissing(12, cursor);
      setMissingItems((prev) => (isMore ? [...prev, ...res.items] : res.items));
      setNextCursor(res.next_cursor);
    } catch {
      // silent
    } finally {
      setLoadingMissing(false);
      setLoadingMore(false);
    }
  }, []);

  useEffect(() => {
    loadMissing();
    api
      .listHomeless()
      .then(setHomelessItems)
      .catch(() => {})
      .finally(() => setLoadingHomeless(false));
  }, [loadMissing]);

  return (
    <>
      <Head title="Traceo" description={t("home.hero.subtitle")} />

      <section className="text-center space-y-3 py-8">
        <h1 className="text-4xl font-bold tracking-tight">{t("home.hero.title")}</h1>
        <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
          {t("home.hero.subtitle")}
        </p>
        <div className="flex justify-center gap-3 pt-2">
          <Link to="/register-choice">
            <Button size="lg">{t("home.cta")}</Button>
          </Link>
        </div>
      </section>

      <section className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Search className="h-5 w-5" />
            {t("missing.listTitle")}
          </h2>
          <Link to="/missing">
            <Button variant="outline" size="sm">{t("common.seeAll")}</Button>
          </Link>
        </div>

        {loadingMissing ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-64 rounded-lg" />
            ))}
          </div>
        ) : (
          <>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
              {missingItems.map((item) => (
                <MissingCard key={item.id} item={item} onSelect={setSelected} />
              ))}
            </div>
            {nextCursor && (
              <div className="flex justify-center">
                <Button variant="outline" onClick={() => loadMissing(nextCursor)} disabled={loadingMore}>
                  {loadingMore ? t("common.loading") : t("missing.loadMore")}
                </Button>
              </div>
            )}
          </>
        )}
      </section>

      <section className="space-y-4 mt-10">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Heart className="h-5 w-5" />
            {t("homeless.title")}
          </h2>
          <Link to="/homeless">
            <Button variant="outline" size="sm">{t("common.seeAll")}</Button>
          </Link>
        </div>

        {loadingHomeless ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-64 rounded-lg" />
            ))}
          </div>
        ) : homelessItems.length === 0 ? (
          <p className="text-muted-foreground">{t("homeless.empty")}</p>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
            {homelessItems.map((item) => (
              <HomelessCard key={item.id} item={item} />
            ))}
          </div>
        )}
      </section>

      {selected && (
        <MissingDetailModal item={selected} onClose={() => setSelected(null)} />
      )}
    </>
  );
}
