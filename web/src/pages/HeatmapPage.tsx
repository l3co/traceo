import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import Head from "@/shared/components/Head";
import MapHeatmap from "@/shared/components/maps/MapHeatmap";
import { api, type LocationPointDTO } from "@/shared/lib/api";

export default function HeatmapPage() {
  const { t } = useTranslation();
  const [points, setPoints] = useState<{ lat: number; lng: number; weight: number }[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getMissingLocations(500)
      .then((res) =>
        setPoints(
          res.locations.map((p: LocationPointDTO) => ({
            lat: p.lat,
            lng: p.lng,
            weight: 1,
          }))
        )
      )
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <>
      <Head title={t("heatmap.title")} description={t("heatmap.description")} />
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">{t("heatmap.title")}</h1>
        <p className="text-muted-foreground text-sm">{t("heatmap.description")}</p>
        {loading ? (
          <div className="h-[500px] flex items-center justify-center text-muted-foreground">
            {t("common.loading")}
          </div>
        ) : (
          <MapHeatmap points={points} className="h-[500px] w-full rounded-lg" />
        )}
      </div>
    </>
  );
}
