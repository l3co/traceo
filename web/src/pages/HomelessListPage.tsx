import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { Plus } from "lucide-react";
import { api, type HomelessResponse } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import HomelessCard from "@/features/homeless/components/HomelessCard";

export default function HomelessListPage() {
  const { t } = useTranslation();
  const [items, setItems] = useState<HomelessResponse[]>([]);
  const [loading, setLoading] = useState(true);

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
        <Link to="/homeless/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            {t("homeless.register")}
          </Button>
        </Link>
      </div>

      {loading ? (
        <p className="text-muted-foreground">{t("common.loading")}</p>
      ) : items.length === 0 ? (
        <p className="text-muted-foreground">{t("homeless.empty")}</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {items.map((item) => (
            <HomelessCard key={item.id} item={item} />
          ))}
        </div>
      )}
    </div>
  );
}
