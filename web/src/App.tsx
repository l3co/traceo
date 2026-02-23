import { useTranslation } from "react-i18next";
import { Search, MapPin, Users, Heart } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { LanguageSwitcher } from "@/shared/components/LanguageSwitcher";
import { HealthStatus } from "@/shared/components/HealthStatus";

function App() {
  const { t } = useTranslation();

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b bg-card">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-4">
          <div className="flex items-center gap-2">
            <Heart className="h-6 w-6 text-rose-500" />
            <span className="text-lg font-bold">
              {t("home.title")}
            </span>
          </div>
          <LanguageSwitcher />
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-4 py-16">
        <section className="mb-16 text-center">
          <h1 className="mb-4 text-4xl font-bold tracking-tight sm:text-5xl">
            {t("home.title")}
          </h1>
          <p className="mb-2 text-xl text-muted-foreground">
            {t("home.subtitle")}
          </p>
          <p className="mx-auto max-w-2xl text-base text-muted-foreground/70">
            {t("home.description")}
          </p>
        </section>

        <section className="mb-16 grid gap-6 sm:grid-cols-3">
          <FeatureCard
            icon={<Search className="h-6 w-6" />}
            title={t("nav.missing")}
          />
          <FeatureCard
            icon={<MapPin className="h-6 w-6" />}
            title={t("nav.sightings")}
          />
          <FeatureCard
            icon={<Users className="h-6 w-6" />}
            title={t("nav.homeless")}
          />
        </section>

        <section className="flex justify-center">
          <HealthStatus />
        </section>
      </main>
    </div>
  );
}

function FeatureCard({ icon, title }: { icon: React.ReactNode; title: string }) {
  return (
    <Card className="transition-colors hover:border-primary/30">
      <CardContent className="flex flex-col items-center gap-3 p-6 text-center">
        <div className="text-muted-foreground">{icon}</div>
        <span className="text-sm font-medium">{title}</span>
      </CardContent>
    </Card>
  );
}

export default App;
