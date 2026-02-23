import { APIProvider } from "@vis.gl/react-google-maps";
import type { ReactNode } from "react";

const API_KEY = import.meta.env.VITE_GOOGLE_MAPS_API_KEY || "";

interface Props {
  children: ReactNode;
}

export default function MapProvider({ children }: Props) {
  if (!API_KEY) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground text-sm p-4">
        Google Maps API key not configured
      </div>
    );
  }

  return <APIProvider apiKey={API_KEY}>{children}</APIProvider>;
}
