import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useAuth } from "@/shared/contexts/AuthContext";
import { api, type UserResponse } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function ProfilePage() {
  const { t } = useTranslation();
  const { user: authUser } = useAuth();

  const [profile, setProfile] = useState<UserResponse | null>(null);
  const [form, setForm] = useState({ name: "", phone: "", cell_phone: "" });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!authUser) return;
    api
      .getUser(authUser.uid)
      .then((u) => {
        setProfile(u);
        setForm({ name: u.name, phone: u.phone || "", cell_phone: u.cell_phone || "" });
      })
      .catch(() => setError(t("profile.loadError")))
      .finally(() => setLoading(false));
  }, [authUser, t]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!authUser) return;
    setMessage("");
    setError("");
    setSaving(true);

    try {
      const updated = await api.updateUser(authUser.uid, form);
      setProfile(updated);
      setMessage(t("profile.success"));
    } catch {
      setError(t("profile.error"));
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">{t("common.loading")}</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-2xl font-bold">{t("profile.title")}</h1>

      <Card>
        <CardHeader>
          <CardTitle>{t("profile.personalInfo")}</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {message && (
              <p className="text-sm text-green-600 text-center">{message}</p>
            )}
            {error && (
              <p className="text-sm text-destructive text-center">{error}</p>
            )}

            <div className="space-y-2">
              <Label>{t("profile.email")}</Label>
              <Input value={profile?.email || ""} disabled />
            </div>

            <div className="space-y-2">
              <Label htmlFor="name">{t("profile.name")}</Label>
              <Input
                id="name"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
                maxLength={150}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="phone">{t("profile.phone")}</Label>
                <Input
                  id="phone"
                  value={form.phone}
                  onChange={(e) => setForm({ ...form, phone: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="cell_phone">{t("profile.cellPhone")}</Label>
                <Input
                  id="cell_phone"
                  value={form.cell_phone}
                  onChange={(e) => setForm({ ...form, cell_phone: e.target.value })}
                />
              </div>
            </div>

            <Button type="submit" disabled={saving}>
              {saving ? t("common.loading") : t("profile.save")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
