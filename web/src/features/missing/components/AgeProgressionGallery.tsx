import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { api } from "@/shared/lib/api";
import { Skeleton } from "@/components/ui/skeleton";

interface Props {
  missingId: string;
  originalPhotoUrl?: string;
}

export default function AgeProgressionGallery({ missingId, originalPhotoUrl }: Props) {
  const { t } = useTranslation();
  const [urls, setUrls] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIndex, setSelectedIndex] = useState(0);

  useEffect(() => {
    api
      .getAgeProgression(missingId)
      .then((res) => setUrls(res.urls))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [missingId]);

  if (loading) {
    return (
      <div className="space-y-2">
        <Skeleton className="h-5 w-48" />
        <Skeleton className="h-48 w-full rounded-lg" />
      </div>
    );
  }

  if (urls.length === 0) {
    return null;
  }

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold">{t("ageProgression.title")}</h3>
      <p className="text-xs text-muted-foreground">{t("ageProgression.disclaimer")}</p>

      <div className="flex gap-4 items-start">
        {originalPhotoUrl && (
          <div className="space-y-1 text-center">
            <img
              src={originalPhotoUrl}
              alt={t("ageProgression.original")}
              className="h-48 w-36 object-cover rounded-lg border"
            />
            <span className="text-xs text-muted-foreground">{t("ageProgression.original")}</span>
          </div>
        )}

        <div className="space-y-1 text-center">
          <img
            src={urls[selectedIndex]}
            alt={t("ageProgression.projection")}
            className="h-48 w-36 object-cover rounded-lg border"
          />
          <span className="text-xs text-muted-foreground">{t("ageProgression.projection")}</span>
        </div>
      </div>

      {urls.length > 1 && (
        <div className="flex gap-2">
          {urls.map((url, i) => (
            <button
              key={url}
              type="button"
              onClick={() => setSelectedIndex(i)}
              className={`h-12 w-12 rounded border overflow-hidden ${
                i === selectedIndex ? "ring-2 ring-primary" : "opacity-60"
              }`}
            >
              <img src={url} alt="" className="h-full w-full object-cover" />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
