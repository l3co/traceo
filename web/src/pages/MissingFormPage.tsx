import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router-dom";
import { useAuth } from "@/shared/contexts/AuthContext";
import { api, type CreateMissingInput } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  GENDER_OPTIONS,
  EYE_OPTIONS,
  HAIR_OPTIONS,
  SKIN_OPTIONS,
} from "@/features/missing/constants";

function SelectField({
  id,
  label,
  value,
  onChange,
  options,
  lang,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: { value: string; labelPt: string; labelEn: string }[];
  lang: string;
}) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <select
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs"
        required
      >
        <option value="">---</option>
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {lang.startsWith("pt") ? o.labelPt : o.labelEn}
          </option>
        ))}
      </select>
    </div>
  );
}

const emptyForm: CreateMissingInput = {
  name: "",
  nickname: "",
  birth_date: "",
  date_of_disappearance: "",
  height: "",
  clothes: "",
  gender: "",
  eyes: "",
  hair: "",
  skin: "",
  lat: 0,
  lng: 0,
  event_report: "",
  tattoo_description: "",
  scar_description: "",
};

export default function MissingFormPage() {
  const { t, i18n } = useTranslation();
  const { user } = useAuth();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = !!id;
  const lang = i18n.language;

  const [form, setForm] = useState<CreateMissingInput>({ ...emptyForm });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [loadingData, setLoadingData] = useState(isEdit);

  useEffect(() => {
    if (!isEdit || !id) return;
    api
      .getMissing(id)
      .then((m) => {
        setForm({
          name: m.name,
          nickname: m.nickname || "",
          birth_date: m.birth_date || "",
          date_of_disappearance: m.date_of_disappearance || "",
          height: m.height || "",
          clothes: m.clothes || "",
          gender: m.gender,
          eyes: m.eyes,
          hair: m.hair,
          skin: m.skin,
          lat: m.lat,
          lng: m.lng,
          event_report: m.event_report || "",
          tattoo_description: m.tattoo_description || "",
          scar_description: m.scar_description || "",
        });
      })
      .catch(() => setError(t("missing.loadError")))
      .finally(() => setLoadingData(false));
  }, [id, isEdit, t]);

  const update = (field: string, value: string | number) =>
    setForm((prev) => ({ ...prev, [field]: value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;
    setError("");
    setLoading(true);

    try {
      if (isEdit && id) {
        await api.updateMissing(id, form);
      } else {
        await api.createMissing(form);
      }
      navigate("/missing");
    } catch (err) {
      setError(err instanceof Error ? err.message : t("missing.saveError"));
    } finally {
      setLoading(false);
    }
  };

  if (loadingData) {
    return (
      <div className="flex justify-center py-12">
        <p className="text-muted-foreground">{t("common.loading")}</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <h1 className="text-2xl font-bold">
        {isEdit ? t("missing.editTitle") : t("missing.createTitle")}
      </h1>

      <Card>
        <CardHeader>
          <CardTitle>{t("missing.personalData")}</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <p className="text-sm text-destructive text-center">{error}</p>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="name">{t("missing.name")}</Label>
                <Input
                  id="name"
                  value={form.name}
                  onChange={(e) => update("name", e.target.value)}
                  required
                  maxLength={200}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="nickname">{t("missing.nickname")}</Label>
                <Input
                  id="nickname"
                  value={form.nickname}
                  onChange={(e) => update("nickname", e.target.value)}
                  maxLength={100}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="birth_date">{t("missing.birthDate")}</Label>
                <Input
                  id="birth_date"
                  placeholder="DD/MM/AAAA"
                  value={form.birth_date}
                  onChange={(e) => update("birth_date", e.target.value)}
                  maxLength={10}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="date_of_disappearance">
                  {t("missing.dateOfDisappearance")}
                </Label>
                <Input
                  id="date_of_disappearance"
                  placeholder="DD/MM/AAAA"
                  value={form.date_of_disappearance}
                  onChange={(e) =>
                    update("date_of_disappearance", e.target.value)
                  }
                  maxLength={10}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <SelectField
                id="gender"
                label={t("missing.gender")}
                value={form.gender}
                onChange={(v) => update("gender", v)}
                options={GENDER_OPTIONS}
                lang={lang}
              />
              <SelectField
                id="eyes"
                label={t("missing.eyes")}
                value={form.eyes}
                onChange={(v) => update("eyes", v)}
                options={EYE_OPTIONS}
                lang={lang}
              />
              <SelectField
                id="hair"
                label={t("missing.hair")}
                value={form.hair}
                onChange={(v) => update("hair", v)}
                options={HAIR_OPTIONS}
                lang={lang}
              />
              <SelectField
                id="skin"
                label={t("missing.skin")}
                value={form.skin}
                onChange={(v) => update("skin", v)}
                options={SKIN_OPTIONS}
                lang={lang}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="height">{t("missing.height")}</Label>
              <Input
                id="height"
                value={form.height}
                onChange={(e) => update("height", e.target.value)}
                placeholder="175cm"
                maxLength={20}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="clothes">{t("missing.clothes")}</Label>
              <textarea
                id="clothes"
                value={form.clothes}
                onChange={(e) => update("clothes", e.target.value)}
                className="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs"
                maxLength={500}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="event_report">{t("missing.eventReport")}</Label>
              <textarea
                id="event_report"
                value={form.event_report}
                onChange={(e) => update("event_report", e.target.value)}
                className="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs"
                maxLength={2000}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="tattoo_description">
                  {t("missing.tattoo")}
                </Label>
                <Input
                  id="tattoo_description"
                  value={form.tattoo_description}
                  onChange={(e) =>
                    update("tattoo_description", e.target.value)
                  }
                  maxLength={500}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="scar_description">{t("missing.scar")}</Label>
                <Input
                  id="scar_description"
                  value={form.scar_description}
                  onChange={(e) => update("scar_description", e.target.value)}
                  maxLength={500}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="lat">{t("missing.latitude")}</Label>
                <Input
                  id="lat"
                  type="number"
                  step="any"
                  value={form.lat}
                  onChange={(e) => update("lat", parseFloat(e.target.value) || 0)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lng">{t("missing.longitude")}</Label>
                <Input
                  id="lng"
                  type="number"
                  step="any"
                  value={form.lng}
                  onChange={(e) => update("lng", parseFloat(e.target.value) || 0)}
                />
              </div>
            </div>

            <div className="flex gap-4">
              <Button type="submit" disabled={loading}>
                {loading
                  ? t("common.loading")
                  : isEdit
                    ? t("missing.save")
                    : t("missing.register")}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate("/missing")}
              >
                {t("common.cancel")}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
