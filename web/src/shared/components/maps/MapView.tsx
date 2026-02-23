import { Map, AdvancedMarker } from "@vis.gl/react-google-maps";
import MapProvider from "./MapProvider";

interface MarkerData {
  lat: number;
  lng: number;
  label?: string;
}

interface Props {
  center?: { lat: number; lng: number };
  zoom?: number;
  markers?: MarkerData[];
  className?: string;
}

const BRAZIL_CENTER = { lat: -15.77, lng: -47.92 };

export default function MapView({
  center = BRAZIL_CENTER,
  zoom = 4,
  markers = [],
  className = "h-[400px] w-full rounded-lg",
}: Props) {
  return (
    <MapProvider>
      <div className={className}>
        <Map
          defaultCenter={center}
          defaultZoom={zoom}
          mapId="traceo-map-view"
          style={{ width: "100%", height: "100%" }}
        >
          {markers.map((m, i) => (
            <AdvancedMarker key={i} position={{ lat: m.lat, lng: m.lng }} />
          ))}
        </Map>
      </div>
    </MapProvider>
  );
}
