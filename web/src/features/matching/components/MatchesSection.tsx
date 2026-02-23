import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { api, type MatchResponse } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";
import { CheckCircle, XCircle } from "lucide-react";

interface Props {
  missingId: string;
}

export default function MatchesSection({ missingId }: Props) {
  const { t } = useTranslation();
  const [matches, setMatches] = useState<MatchResponse[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchMatches = () => {
    api
      .getMissingMatches(missingId)
      .then(setMatches)
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchMatches();
  }, [missingId]);

  const handleUpdateStatus = async (matchId: string, status: string) => {
    try {
      await api.updateMatchStatus(matchId, status);
      fetchMatches();
    } catch {
      // silent
    }
  };

  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("common.loading")}</p>;
  }

  if (matches.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("matches.noMatches")}</p>;
  }

  return (
    <div className="space-y-3">
      {matches.map((match) => (
        <div
          key={match.id}
          className="rounded-lg border p-3 space-y-2"
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div
                className="h-2 w-2 rounded-full"
                style={{
                  backgroundColor:
                    match.score >= 0.8
                      ? "#22c55e"
                      : match.score >= 0.6
                        ? "#f59e0b"
                        : "#ef4444",
                }}
              />
              <span className="text-sm font-medium">
                {(match.score * 100).toFixed(0)}% {t("matches.similarity")}
              </span>
            </div>
            <span className="text-xs text-muted-foreground">
              {match.status === "pending"
                ? t("matches.pending")
                : match.status === "confirmed"
                  ? t("matches.confirmed")
                  : t("matches.rejected")}
            </span>
          </div>

          {match.gemini_analysis && (
            <p className="text-xs text-muted-foreground line-clamp-2">
              {match.gemini_analysis}
            </p>
          )}

          {match.status === "pending" && (
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="default"
                className="bg-green-600 hover:bg-green-700"
                onClick={() => handleUpdateStatus(match.id, "confirmed")}
              >
                <CheckCircle className="mr-1 h-3 w-3" />
                {t("matches.confirm")}
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleUpdateStatus(match.id, "rejected")}
              >
                <XCircle className="mr-1 h-3 w-3" />
                {t("matches.reject")}
              </Button>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
