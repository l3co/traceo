import { useEffect } from "react";
import { Map, useMap } from "@vis.gl/react-google-maps";
import MapProvider from "./MapProvider";

interface HeatmapPoint {
  lat: number;
  lng: number;
  weight?: number;
}

interface Props {
  points: HeatmapPoint[];
  className?: string;
}

const BRAZIL_CENTER = { lat: -15.77, lng: -47.92 };

function HeatmapLayer({ points }: { points: HeatmapPoint[] }) {
  const map = useMap();

  useEffect(() => {
    if (!map || !points.length) return;

    const heatmap = new google.maps.visualization.HeatmapLayer({
      data: points.map((p) => ({
        location: new google.maps.LatLng(p.lat, p.lng),
        weight: p.weight ?? 1,
      })),
      radius: 30,
      opacity: 0.7,
    });
    heatmap.setMap(map);

    return () => {
      heatmap.setMap(null);
    };
  }, [map, points]);

  return null;
}

export default function MapHeatmap({ points, className = "h-[500px] w-full rounded-lg" }: Props) {
  return (
    <MapProvider>
      <div className={className}>
        <Map
          defaultCenter={BRAZIL_CENTER}
          defaultZoom={4}
          mapId="traceo-heatmap"
          style={{ width: "100%", height: "100%" }}
        >
          <HeatmapLayer points={points} />
        </Map>
      </div>
    </MapProvider>
  );
}
