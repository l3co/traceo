import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Activity, CheckCircle, XCircle, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { api } from "@/shared/lib/api";

interface HealthData {
  status: string;
  message: string;
  uptime: string;
}

type ConnectionState = "checking" | "connected" | "disconnected";

export function HealthStatus() {
  const { t } = useTranslation();
  const [state, setState] = useState<ConnectionState>("checking");
  const [data, setData] = useState<HealthData | null>(null);

  const check = async () => {
    setState("checking");
    try {
      const result = await api.health();
      setData(result);
      setState("connected");
    } catch {
      setState("disconnected");
      setData(null);
    }
  };

  useEffect(() => {
    check();
  }, []);

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="flex flex-row items-center gap-2 space-y-0 pb-4">
        <Activity className="h-5 w-5 text-muted-foreground" />
        <CardTitle className="text-lg">{t("health.title")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            {t("health.status")}
          </span>
          <StatusBadge state={state} />
        </div>

        {data && (
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">
              {t("health.uptime")}
            </span>
            <span className="text-sm font-medium">
              {data.uptime}
            </span>
          </div>
        )}

        {state === "disconnected" && (
          <Button onClick={check} className="mt-2 w-full">
            {t("common.retry")}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

function StatusBadge({ state }: { state: ConnectionState }) {
  const { t } = useTranslation();

  switch (state) {
    case "checking":
      return (
        <Badge variant="secondary" className="gap-1.5">
          <Loader2 className="h-3 w-3 animate-spin" />
          {t("health.checking")}
        </Badge>
      );
    case "connected":
      return (
        <Badge className="gap-1.5 border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400">
          <CheckCircle className="h-3 w-3" />
          {t("health.connected")}
        </Badge>
      );
    case "disconnected":
      return (
        <Badge variant="destructive" className="gap-1.5">
          <XCircle className="h-3 w-3" />
          {t("health.disconnected")}
        </Badge>
      );
  }
}
