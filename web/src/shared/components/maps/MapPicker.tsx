import { useState, useCallback } from "react";
import { Map, AdvancedMarker, type MapMouseEvent } from "@vis.gl/react-google-maps";
import { useTranslation } from "react-i18next";
import { MapPin } from "lucide-react";
import { Button } from "@/components/ui/button";
import MapProvider from "./MapProvider";

interface Props {
  value?: { lat: number; lng: number };
  onChange: (pos: { lat: number; lng: number }) => void;
  className?: string;
}

const BRAZIL_CENTER = { lat: -15.77, lng: -47.92 };

export default function MapPicker({ value, onChange, className = "h-[300px] w-full rounded-lg" }: Props) {
  const { t } = useTranslation();
  const [position, setPosition] = useState(value);

  const handleClick = useCallback(
    (e: MapMouseEvent) => {
      const detail = e.detail;
      if (detail.latLng) {
        const pos = { lat: detail.latLng.lat, lng: detail.latLng.lng };
        setPosition(pos);
        onChange(pos);
      }
    },
    [onChange],
  );

  const handleUseMyLocation = () => {
    if (!navigator.geolocation) return;
    navigator.geolocation.getCurrentPosition(
      (geo) => {
        const pos = { lat: geo.coords.latitude, lng: geo.coords.longitude };
        setPosition(pos);
        onChange(pos);
      },
      () => {},
    );
  };

  return (
    <MapProvider>
      <div className="space-y-2">
        <div className={className}>
          <Map
            defaultCenter={position || BRAZIL_CENTER}
            defaultZoom={position ? 14 : 4}
            mapId="traceo-map-picker"
            style={{ width: "100%", height: "100%" }}
            onClick={handleClick}
          >
            {position && <AdvancedMarker position={position} />}
          </Map>
        </div>
        <Button type="button" variant="outline" size="sm" onClick={handleUseMyLocation}>
          <MapPin className="mr-2 h-4 w-4" />
          {t("sighting.useMyLocation")}
        </Button>
      </div>
    </MapProvider>
  );
}
