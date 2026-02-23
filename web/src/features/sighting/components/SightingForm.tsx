import { useState } from "react";
import { useTranslation } from "react-i18next";
import { MapPin, Send } from "lucide-react";
import { api } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface Props {
  missingId: string;
  missingName: string;
  onSuccess: () => void;
  onCancel: () => void;
}

export default function SightingForm({ missingId, missingName, onSuccess, onCancel }: Props) {
  const { t } = useTranslation();
  const [lat, setLat] = useState("");
  const [lng, setLng] = useState("");
  const [observation, setObservation] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [geoLoading, setGeoLoading] = useState(false);

  const handleGetLocation = () => {
    if (!navigator.geolocation) return;
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLat(String(pos.coords.latitude));
        setLng(String(pos.coords.longitude));
        setGeoLoading(false);
      },
      () => setGeoLoading(false)
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await api.createSighting(missingId, {
        lat: parseFloat(lat),
        lng: parseFloat(lng),
        observation,
      });
      onSuccess();
    } catch {
      setError(t("sighting.submitError"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <p className="text-sm text-muted-foreground">
        {t("sighting.formDescription", { name: missingName })}
      </p>

      <div className="space-y-2">
        <Label>{t("sighting.observation")}</Label>
        <textarea
          className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          value={observation}
          onChange={(e) => setObservation(e.target.value)}
          placeholder={t("sighting.observationPlaceholder")}
          required
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{t("missing.lat")}</Label>
          <Input
            type="number"
            step="any"
            value={lat}
            onChange={(e) => setLat(e.target.value)}
            required
          />
        </div>
        <div className="space-y-2">
          <Label>{t("missing.lng")}</Label>
          <Input
            type="number"
            step="any"
            value={lng}
            onChange={(e) => setLng(e.target.value)}
            required
          />
        </div>
      </div>

      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={handleGetLocation}
        disabled={geoLoading}
      >
        <MapPin className="mr-2 h-4 w-4" />
        {geoLoading ? t("common.loading") : t("sighting.useMyLocation")}
      </Button>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="flex justify-end gap-2">
        <Button type="button" variant="ghost" onClick={onCancel}>
          {t("common.cancel")}
        </Button>
        <Button type="submit" disabled={loading}>
          <Send className="mr-2 h-4 w-4" />
          {loading ? t("common.loading") : t("sighting.submit")}
        </Button>
      </div>
    </form>
  );
}
