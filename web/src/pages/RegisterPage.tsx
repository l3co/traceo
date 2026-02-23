import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "@/shared/contexts/AuthContext";
import { api } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function RegisterPage() {
  const { t } = useTranslation();
  const { signIn } = useAuth();
  const navigate = useNavigate();

  const [form, setForm] = useState({
    name: "",
    email: "",
    password: "",
    phone: "",
    cell_phone: "",
    accepted_terms: false,
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const update = (field: string, value: string | boolean) =>
    setForm((prev) => ({ ...prev, [field]: value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await api.createUser(form);
      await signIn(form.email, form.password);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : t("register.error"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 px-4 py-8">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">{t("register.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <p className="text-sm text-destructive text-center">{error}</p>
            )}

            <div className="space-y-2">
              <Label htmlFor="name">{t("register.name")}</Label>
              <Input
                id="name"
                value={form.name}
                onChange={(e) => update("name", e.target.value)}
                required
                maxLength={150}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">{t("register.email")}</Label>
              <Input
                id="email"
                type="email"
                value={form.email}
                onChange={(e) => update("email", e.target.value)}
                required
                autoComplete="email"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">{t("register.password")}</Label>
              <Input
                id="password"
                type="password"
                value={form.password}
                onChange={(e) => update("password", e.target.value)}
                required
                minLength={6}
                autoComplete="new-password"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="phone">{t("register.phone")}</Label>
                <Input
                  id="phone"
                  value={form.phone}
                  onChange={(e) => update("phone", e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="cell_phone">{t("register.cellPhone")}</Label>
                <Input
                  id="cell_phone"
                  value={form.cell_phone}
                  onChange={(e) => update("cell_phone", e.target.value)}
                />
              </div>
            </div>

            <div className="flex items-center gap-2">
              <input
                id="terms"
                type="checkbox"
                checked={form.accepted_terms}
                onChange={(e) => update("accepted_terms", e.target.checked)}
                required
                className="h-4 w-4 rounded border-input"
              />
              <Label htmlFor="terms" className="text-sm font-normal">
                {t("register.acceptTerms")}
              </Label>
            </div>

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? t("common.loading") : t("register.submit")}
            </Button>

            <p className="text-center text-sm text-muted-foreground">
              {t("register.hasAccount")}{" "}
              <Link to="/login" className="text-primary hover:underline">
                {t("register.loginLink")}
              </Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
