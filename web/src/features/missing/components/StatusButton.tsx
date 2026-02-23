import { useState } from "react";
import { useTranslation } from "react-i18next";
import { CheckCircle, Search } from "lucide-react";
import { api } from "@/shared/lib/api";
import { Button } from "@/components/ui/button";

interface Props {
  missingId: string;
  currentStatus: string;
  isOwner: boolean;
  onStatusChanged: (newStatus: string) => void;
}

export default function StatusButton({ missingId, currentStatus, isOwner, onStatusChanged }: Props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);

  if (!isOwner) return null;

  const isFound = currentStatus === "found";
  const newStatus = isFound ? "disappeared" : "found";

  const handleClick = async () => {
    if (!confirm(t(isFound ? "status.confirmReactivate" : "status.confirmFound"))) return;
    setLoading(true);
    try {
      await api.updateMissingStatus(missingId, newStatus);
      onStatusChanged(newStatus);
    } catch {
      // silent
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button
      variant={isFound ? "outline" : "default"}
      size="sm"
      onClick={handleClick}
      disabled={loading}
      className={isFound ? "" : "bg-green-600 hover:bg-green-700"}
    >
      {isFound ? (
        <>
          <Search className="mr-2 h-4 w-4" />
          {t("status.reactivate")}
        </>
      ) : (
        <>
          <CheckCircle className="mr-2 h-4 w-4" />
          {t("status.markFound")}
        </>
      )}
    </Button>
  );
}
