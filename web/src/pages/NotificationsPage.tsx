import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import Head from "@/shared/components/Head";
import MapView from "@/shared/components/maps/MapView";
import { api, type SightingResponse } from "@/shared/lib/api";
import { useAuth } from "@/shared/contexts/AuthContext";

export default function NotificationsPage() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [sightings, setSightings] = useState<SightingResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [missingIds, setMissingIds] = useState<string[]>([]);

  useEffect(() => {
    if (!user?.uid) return;
    api
      .getUserMissing(user.uid)
      .then(async (missingList) => {
        const ids = missingList.map((m) => m.id);
        setMissingIds(ids);
        const allSightings: SightingResponse[] = [];
        for (const id of ids) {
          try {
            const s = await api.getSightings(id);
            allSightings.push(...s);
          } catch {
            // skip
          }
        }
        allSightings.sort(
          (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );
        setSightings(allSightings);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [user?.uid]);

  const markers = sightings.map((s) => ({
    lat: s.lat,
    lng: s.lng,
    label: s.observation,
  }));

  return (
    <>
      <Head title={t("notifications.title")} />
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">{t("notifications.title")}</h1>

        {loading ? (
          <p className="text-muted-foreground">{t("common.loading")}</p>
        ) : missingIds.length === 0 ? (
          <p className="text-muted-foreground">{t("notifications.noMissing")}</p>
        ) : sightings.length === 0 ? (
          <p className="text-muted-foreground">{t("notifications.noSightings")}</p>
        ) : (
          <>
            <MapView markers={markers} className="h-[400px] w-full rounded-lg" />
            <div className="space-y-2">
              {sightings.map((s) => (
                <div key={s.id} className="rounded-lg border p-3 space-y-1">
                  <p className="text-sm">{s.observation}</p>
                  <p className="text-xs text-muted-foreground">
                    {new Date(s.created_at).toLocaleDateString()}
                  </p>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </>
  );
}
