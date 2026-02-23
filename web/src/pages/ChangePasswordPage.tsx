import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useAuth } from "@/shared/contexts/AuthContext";
import { api } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function ChangePasswordPage() {
  const { t } = useTranslation();
  const { user } = useAuth();

  const [newPassword, setNewPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage("");
    setError("");

    if (newPassword !== confirm) {
      setError(t("password.mismatch"));
      return;
    }

    if (!user) return;
    setLoading(true);

    try {
      await api.changePassword(user.uid, newPassword);
      setMessage(t("password.success"));
      setNewPassword("");
      setConfirm("");
    } catch {
      setError(t("password.error"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-md space-y-6">
      <h1 className="text-2xl font-bold">{t("password.title")}</h1>

      <Card>
        <CardHeader>
          <CardTitle>{t("password.change")}</CardTitle>
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
              <Label htmlFor="newPassword">{t("password.new")}</Label>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
                minLength={6}
                autoComplete="new-password"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="confirm">{t("password.confirm")}</Label>
              <Input
                id="confirm"
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                required
                minLength={6}
                autoComplete="new-password"
              />
            </div>

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? t("common.loading") : t("password.submit")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
