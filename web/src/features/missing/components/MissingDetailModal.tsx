import { useTranslation } from "react-i18next";
import { X, MapPin, Calendar, User, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import type { MissingResponse } from "@/shared/lib/api";
import {
  getLabel,
  GENDER_OPTIONS,
  EYE_OPTIONS,
  HAIR_OPTIONS,
  SKIN_OPTIONS,
  STATUS_OPTIONS,
} from "../constants";

interface MissingDetailModalProps {
  item: MissingResponse;
  onClose: () => void;
}

export default function MissingDetailModal({
  item,
  onClose,
}: MissingDetailModalProps) {
  const { t, i18n } = useTranslation();
  const lang = i18n.language;

  const isFound = item.status === "found";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="relative w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-lg bg-background shadow-xl">
        <div className="sticky top-0 z-10 flex items-center justify-between border-b bg-background px-6 py-4">
          <h2 className="text-xl font-bold">{item.name}</h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>

        <div className="p-6 space-y-6">
          <div className="flex flex-col sm:flex-row gap-6">
            <div className="w-full sm:w-1/3">
              {item.photo_url ? (
                <img
                  src={item.photo_url}
                  alt={item.name}
                  className="w-full rounded-lg object-cover aspect-[3/4]"
                />
              ) : (
                <div className="flex aspect-[3/4] items-center justify-center rounded-lg bg-muted">
                  <User className="h-20 w-20 text-muted-foreground/30" />
                </div>
              )}
            </div>

            <div className="flex-1 space-y-3">
              <div className="flex items-center gap-2">
                <span
                  className={`rounded-full px-3 py-1 text-sm font-medium ${
                    isFound
                      ? "bg-green-100 text-green-800"
                      : "bg-red-100 text-red-800"
                  }`}
                >
                  {getLabel(STATUS_OPTIONS, item.status, lang)}
                </span>
                {item.was_child && (
                  <span className="rounded-full bg-yellow-100 px-3 py-1 text-sm font-medium text-yellow-800">
                    {t("missing.wasChild")}
                  </span>
                )}
              </div>

              {item.nickname && (
                <p className="text-muted-foreground">
                  {t("missing.nickname")}: {item.nickname}
                </p>
              )}

              {item.age > 0 && (
                <p>
                  {t("missing.age")}: {item.age} {t("missing.years")}
                </p>
              )}

              {item.date_of_disappearance && (
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span>
                    {t("missing.disappearedOn")}: {item.date_of_disappearance}
                  </span>
                </div>
              )}

              {item.birth_date && (
                <p className="text-sm text-muted-foreground">
                  {t("missing.birthDate")}: {item.birth_date}
                </p>
              )}
            </div>
          </div>

          <Separator />

          <div>
            <h3 className="font-semibold mb-3">
              {t("missing.physicalTraits")}
            </h3>
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>
                <span className="text-muted-foreground">
                  {t("missing.gender")}:
                </span>{" "}
                {getLabel(GENDER_OPTIONS, item.gender, lang)}
              </div>
              <div>
                <span className="text-muted-foreground">
                  {t("missing.eyes")}:
                </span>{" "}
                {getLabel(EYE_OPTIONS, item.eyes, lang)}
              </div>
              <div>
                <span className="text-muted-foreground">
                  {t("missing.hair")}:
                </span>{" "}
                {getLabel(HAIR_OPTIONS, item.hair, lang)}
              </div>
              <div>
                <span className="text-muted-foreground">
                  {t("missing.skin")}:
                </span>{" "}
                {getLabel(SKIN_OPTIONS, item.skin, lang)}
              </div>
              {item.height && (
                <div>
                  <span className="text-muted-foreground">
                    {t("missing.height")}:
                  </span>{" "}
                  {item.height}
                </div>
              )}
            </div>
          </div>

          {item.clothes && (
            <>
              <Separator />
              <div>
                <h3 className="font-semibold mb-1">{t("missing.clothes")}</h3>
                <p className="text-sm">{item.clothes}</p>
              </div>
            </>
          )}

          {(item.has_tattoo || item.has_scar) && (
            <>
              <Separator />
              <div className="space-y-2">
                <h3 className="font-semibold">
                  {t("missing.distinguishingMarks")}
                </h3>
                {item.tattoo_description && (
                  <p className="text-sm">
                    <span className="font-medium">
                      {t("missing.tattoo")}:
                    </span>{" "}
                    {item.tattoo_description}
                  </p>
                )}
                {item.scar_description && (
                  <p className="text-sm">
                    <span className="font-medium">{t("missing.scar")}:</span>{" "}
                    {item.scar_description}
                  </p>
                )}
              </div>
            </>
          )}

          {item.event_report && (
            <>
              <Separator />
              <div className="flex items-start gap-2">
                <FileText className="h-4 w-4 mt-0.5 text-muted-foreground" />
                <div>
                  <h3 className="font-semibold mb-1">
                    {t("missing.eventReport")}
                  </h3>
                  <p className="text-sm">{item.event_report}</p>
                </div>
              </div>
            </>
          )}

          {(item.lat !== 0 || item.lng !== 0) && (
            <>
              <Separator />
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <MapPin className="h-4 w-4" />
                <span>
                  {t("missing.location")}: {item.lat.toFixed(4)},{" "}
                  {item.lng.toFixed(4)}
                </span>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
