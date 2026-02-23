import { useTranslation } from "react-i18next";
import { User, MapPin } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import type { HomelessResponse } from "@/shared/lib/api";
import { getLabel, GENDER_OPTIONS } from "@/features/missing/constants";

interface Props {
  item: HomelessResponse;
}

export default function HomelessCard({ item }: Props) {
  const { i18n } = useTranslation();
  const lang = i18n.language;

  return (
    <Card className="overflow-hidden">
      <div className="aspect-[3/4] relative">
        {item.photo_url ? (
          <img
            src={item.photo_url}
            alt={item.name}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center bg-muted">
            <User className="h-16 w-16 text-muted-foreground/30" />
          </div>
        )}
      </div>
      <CardContent className="p-4 space-y-1">
        <h3 className="font-semibold truncate">{item.name}</h3>
        {item.nickname && (
          <p className="text-sm text-muted-foreground truncate">
            "{item.nickname}"
          </p>
        )}
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <span>{getLabel(GENDER_OPTIONS, item.gender, lang)}</span>
          {item.age > 0 && <span>· {item.age} anos</span>}
        </div>
        {(item.lat !== 0 || item.lng !== 0) && (
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <MapPin className="h-3 w-3" />
            <span>{item.lat.toFixed(2)}, {item.lng.toFixed(2)}</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
