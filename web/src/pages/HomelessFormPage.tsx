import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { MapPin } from "lucide-react";
import { api } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  getLabel,
  GENDER_OPTIONS,
  EYE_OPTIONS,
  HAIR_OPTIONS,
  SKIN_OPTIONS,
} from "@/features/missing/constants";

export default function HomelessFormPage() {
  const { t, i18n } = useTranslation();
  const lang = i18n.language;
  const navigate = useNavigate();

  const [form, setForm] = useState({
    name: "",
    nickname: "",
    birth_date: "",
    gender: "",
    eyes: "",
    hair: "",
    skin: "",
    photo_url: "",
    lat: "",
    lng: "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [geoLoading, setGeoLoading] = useState(false);

  const set = (field: string, value: string) =>
    setForm((prev) => ({ ...prev, [field]: value }));

  const handleGetLocation = () => {
    if (!navigator.geolocation) return;
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        set("lat", String(pos.coords.latitude));
        set("lng", String(pos.coords.longitude));
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
      await api.createHomeless({
        name: form.name,
        nickname: form.nickname || undefined,
        birth_date: form.birth_date || undefined,
        gender: form.gender,
        eyes: form.eyes,
        hair: form.hair,
        skin: form.skin,
        photo_url: form.photo_url || undefined,
        lat: parseFloat(form.lat) || 0,
        lng: parseFloat(form.lng) || 0,
      });
      navigate("/homeless");
    } catch {
      setError(t("homeless.saveError"));
    } finally {
      setLoading(false);
    }
  };

  const selectClass =
    "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2";

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-2xl font-bold">{t("homeless.formTitle")}</h1>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>{t("missing.name")} *</Label>
            <Input
              value={form.name}
              onChange={(e) => set("name", e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>{t("missing.nickname")}</Label>
            <Input
              value={form.nickname}
              onChange={(e) => set("nickname", e.target.value)}
            />
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>{t("missing.birthDate")}</Label>
            <Input
              value={form.birth_date}
              onChange={(e) => set("birth_date", e.target.value)}
              placeholder="DD/MM/YYYY"
            />
          </div>
          <div className="space-y-2">
            <Label>{t("missing.photoUrl")}</Label>
            <Input
              value={form.photo_url}
              onChange={(e) => set("photo_url", e.target.value)}
              placeholder="https://..."
            />
          </div>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div className="space-y-2">
            <Label>{t("missing.gender")} *</Label>
            <select
              className={selectClass}
              value={form.gender}
              onChange={(e) => set("gender", e.target.value)}
              required
            >
              <option value="">--</option>
              {GENDER_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {getLabel(GENDER_OPTIONS, o.value, lang)}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label>{t("missing.eyes")} *</Label>
            <select
              className={selectClass}
              value={form.eyes}
              onChange={(e) => set("eyes", e.target.value)}
              required
            >
              <option value="">--</option>
              {EYE_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {getLabel(EYE_OPTIONS, o.value, lang)}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label>{t("missing.hair")} *</Label>
            <select
              className={selectClass}
              value={form.hair}
              onChange={(e) => set("hair", e.target.value)}
              required
            >
              <option value="">--</option>
              {HAIR_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {getLabel(HAIR_OPTIONS, o.value, lang)}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label>{t("missing.skin")} *</Label>
            <select
              className={selectClass}
              value={form.skin}
              onChange={(e) => set("skin", e.target.value)}
              required
            >
              <option value="">--</option>
              {SKIN_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {getLabel(SKIN_OPTIONS, o.value, lang)}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="space-y-2">
          <Label>{t("missing.location")}</Label>
          <div className="grid grid-cols-2 gap-4">
            <Input
              type="number"
              step="any"
              placeholder={t("missing.lat")}
              value={form.lat}
              onChange={(e) => set("lat", e.target.value)}
            />
            <Input
              type="number"
              step="any"
              placeholder={t("missing.lng")}
              value={form.lng}
              onChange={(e) => set("lng", e.target.value)}
            />
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
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}

        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="ghost"
            onClick={() => navigate("/homeless")}
          >
            {t("common.cancel")}
          </Button>
          <Button type="submit" disabled={loading}>
            {loading ? t("common.loading") : t("common.save")}
          </Button>
        </div>
      </form>
    </div>
  );
}
