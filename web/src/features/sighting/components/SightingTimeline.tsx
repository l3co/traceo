import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Eye, MapPin } from "lucide-react";
import { api, type SightingResponse } from "@/shared/lib/api";

interface Props {
  missingId: string;
}

export default function SightingTimeline({ missingId }: Props) {
  const { t } = useTranslation();
  const [sightings, setSightings] = useState<SightingResponse[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getSightings(missingId)
      .then(setSightings)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [missingId]);

  if (loading) {
    return (
      <p className="text-sm text-muted-foreground">{t("common.loading")}</p>
    );
  }

  if (sightings.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("sighting.noSightings")}
      </p>
    );
  }

  return (
    <div className="space-y-3">
      <h4 className="flex items-center gap-2 text-sm font-semibold">
        <Eye className="h-4 w-4" />
        {t("sighting.title")} ({sightings.length})
      </h4>
      <div className="space-y-2">
        {sightings.map((s) => (
          <div
            key={s.id}
            className="rounded-md border p-3 text-sm space-y-1"
          >
            <p>{s.observation}</p>
            <div className="flex items-center gap-4 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <MapPin className="h-3 w-3" />
                {s.lat.toFixed(4)}, {s.lng.toFixed(4)}
              </span>
              <span>
                {new Date(s.created_at).toLocaleDateString()}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
