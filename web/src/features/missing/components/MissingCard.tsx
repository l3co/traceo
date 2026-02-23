import { useTranslation } from "react-i18next";
import { MapPin, Calendar, User } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import type { MissingResponse } from "@/shared/lib/api";
import { getLabel, GENDER_OPTIONS, STATUS_OPTIONS } from "../constants";

interface MissingCardProps {
  item: MissingResponse;
  onSelect: (item: MissingResponse) => void;
}

export default function MissingCard({ item, onSelect }: MissingCardProps) {
  const { t, i18n } = useTranslation();
  const lang = i18n.language;

  const isFound = item.status === "found";

  return (
    <Card
      className="overflow-hidden cursor-pointer hover:shadow-md transition-shadow"
      onClick={() => onSelect(item)}
    >
      <div className="relative h-48 bg-muted">
        {item.photo_url ? (
          <img
            src={item.photo_url}
            alt={item.name}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full items-center justify-center">
            <User className="h-16 w-16 text-muted-foreground/30" />
          </div>
        )}
        <span
          className={`absolute top-2 right-2 rounded-full px-2 py-0.5 text-xs font-medium ${
            isFound
              ? "bg-green-100 text-green-800"
              : "bg-red-100 text-red-800"
          }`}
        >
          {getLabel(STATUS_OPTIONS, item.status, lang)}
        </span>
      </div>
      <CardContent className="p-4 space-y-2">
        <h3 className="font-semibold text-base truncate">{item.name}</h3>

        <div className="flex items-center gap-1 text-sm text-muted-foreground">
          <User className="h-3.5 w-3.5" />
          <span>
            {getLabel(GENDER_OPTIONS, item.gender, lang)}
            {item.age > 0 && ` • ${item.age} ${t("missing.years")}`}
          </span>
        </div>

        {item.date_of_disappearance && (
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <Calendar className="h-3.5 w-3.5" />
            <span>{item.date_of_disappearance}</span>
          </div>
        )}

        {(item.lat !== 0 || item.lng !== 0) && (
          <div className="flex items-center gap-1 text-sm text-muted-foreground">
            <MapPin className="h-3.5 w-3.5" />
            <span>{t("missing.hasLocation")}</span>
          </div>
        )}

        <Button variant="outline" size="sm" className="w-full mt-2">
          {t("missing.details")}
        </Button>
      </CardContent>
    </Card>
  );
}
