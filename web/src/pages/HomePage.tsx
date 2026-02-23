import { useTranslation } from "react-i18next";
import { HealthStatus } from "@/shared/components/HealthStatus";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function HomePage() {
  const { t } = useTranslation();

  const features = [
    { key: "home.features.search" },
    { key: "home.features.register" },
    { key: "home.features.map" },
    { key: "home.features.notifications" },
  ];

  return (
    <div className="space-y-8">
      <section className="text-center space-y-2">
        <h2 className="text-3xl font-bold">{t("home.hero.title")}</h2>
        <p className="text-muted-foreground text-lg">
          {t("home.hero.subtitle")}
        </p>
      </section>

      <section>
        <h3 className="text-xl font-semibold mb-4">
          {t("home.features.title")}
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {features.map((f) => (
            <Card key={f.key}>
              <CardHeader>
                <CardTitle className="text-base">
                  {t(`${f.key}.title`)}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  {t(`${f.key}.description`)}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <section>
        <HealthStatus />
      </section>
    </div>
  );
}
