import { useRef, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { MapPin } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export interface LocationData {
  address: string;
  lat: number;
  lng: number;
}

interface Props {
  value?: LocationData;
  onChange: (location: LocationData) => void;
}

export default function AddressInput({ value, onChange }: Props) {
  const { t } = useTranslation();
  const inputRef = useRef<HTMLInputElement>(null);
  const autocompleteRef = useRef<google.maps.places.Autocomplete | null>(null);
  const [address, setAddress] = useState(value?.address || "");
  const [geoLoading, setGeoLoading] = useState(false);

  useEffect(() => {
    if (!inputRef.current || !window.google?.maps?.places) return;
    if (autocompleteRef.current) return;

    const autocomplete = new google.maps.places.Autocomplete(inputRef.current, {
      types: ["geocode", "establishment"],
      componentRestrictions: { country: "br" },
      fields: ["formatted_address", "geometry"],
    });

    autocomplete.addListener("place_changed", () => {
      const place = autocomplete.getPlace();
      if (place.geometry?.location) {
        const loc: LocationData = {
          address: place.formatted_address || "",
          lat: place.geometry.location.lat(),
          lng: place.geometry.location.lng(),
        };
        setAddress(loc.address);
        onChange(loc);
      }
    });

    autocompleteRef.current = autocomplete;
  }, [onChange]);

  const handleUseMyLocation = () => {
    if (!navigator.geolocation) return;
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const lat = pos.coords.latitude;
        const lng = pos.coords.longitude;

        if (window.google?.maps) {
          const geocoder = new google.maps.Geocoder();
          geocoder.geocode({ location: { lat, lng } }, (results, status) => {
            const addr = status === "OK" && results?.[0]
              ? results[0].formatted_address
              : `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
            setAddress(addr);
            onChange({ address: addr, lat, lng });
            setGeoLoading(false);
          });
        } else {
          const addr = `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
          setAddress(addr);
          onChange({ address: addr, lat, lng });
          setGeoLoading(false);
        }
      },
      () => setGeoLoading(false),
    );
  };

  return (
    <div className="space-y-2">
      <Input
        ref={inputRef}
        value={address}
        onChange={(e) => setAddress(e.target.value)}
        placeholder={t("location.placeholder")}
      />
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={handleUseMyLocation}
        disabled={geoLoading}
      >
        <MapPin className="mr-2 h-4 w-4" />
        {geoLoading ? t("common.loading") : t("location.useMyLocation")}
      </Button>
    </div>
  );
}
